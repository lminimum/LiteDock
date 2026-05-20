// Package image implements Docker image management business logic.
package image

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/filters"
	dockerImage "github.com/docker/docker/api/types/image"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/repo"
	"github.com/lminimum/LiteDock/pkg/dockerclient"
	pkgErrors "github.com/lminimum/LiteDock/pkg/errors"
	"github.com/lminimum/LiteDock/pkg/logger"
	"github.com/lminimum/LiteDock/pkg/sshclient"
)

const (
	localMachineID   = "local"
	localMachineHost = "localhost"
)

// ImageUseCase implements Docker image management business logic.
type ImageUseCase struct {
	imageRepo         repo.ImageRepo
	remoteMachineRepo repo.RemoteMachineRepo
	cacheMaxAge       time.Duration
	l                 logger.Interface

	// testDockerClient is a test hook for injecting a mock dockerclient.Client.
	// It is nil in production and set only in tests.
	testDockerClient dockerclient.Client
}

// NewImageUseCase creates a new ImageUseCase.
func NewImageUseCase(imageRepo repo.ImageRepo, rmRepo repo.RemoteMachineRepo, cacheMaxAge time.Duration, l logger.Interface) *ImageUseCase {
	return &ImageUseCase{
		imageRepo:         imageRepo,
		remoteMachineRepo: rmRepo,
		cacheMaxAge:       cacheMaxAge,
		l:                 l,
	}
}

// List returns images for a machine with a cache-first strategy.
// On first call (empty cache), it fetches from Docker and caches the result.
// Subsequent calls return cached data while triggering an async refresh if the cache is stale.
func (uc *ImageUseCase) List(ctx context.Context, machineID string) ([]entity.Image, error) {
	images, err := uc.imageRepo.ListByMachine(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase.List - imageRepo.ListByMachine: %w", err)
	}

	valid, err := uc.imageRepo.IsCacheValid(ctx, machineID, uc.cacheMaxAge)
	if err != nil {
		uc.l.Warn("ImageUseCase.List - imageRepo.IsCacheValid: %v", err)
	}

	if len(images) == 0 {
		return uc.fetchImagesFromDocker(ctx, machineID)
	}

	if !valid {
		go uc.refresh(machineID)
	}

	return images, nil
}

// Inspect returns detailed information about a Docker image.
func (uc *ImageUseCase) Inspect(ctx context.Context, machineID, imageID string) (*entity.Image, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase.Inspect - getDockerClient: %w", err)
	}
	defer cli.Close()

	resp, err := cli.ImageInspect(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase.Inspect - cli.ImageInspect: %w", err)
	}

	img := inspectToEntity(&resp, machineID)
	return &img, nil
}

// Pull pulls a Docker image on the specified machine.
func (uc *ImageUseCase) Pull(ctx context.Context, machineID, repository, tag string) (*entity.Image, error) {
	if repository == "" {
		return nil, fmt.Errorf("ImageUseCase.Pull - repository is required")
	}

	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase.Pull - getDockerClient: %w", err)
	}
	defer cli.Close()

	ref := repository
	if tag != "" {
		ref = repository + ":" + tag
	}

	err = cli.ImagePull(ctx, ref, dockerImage.PullOptions{})
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase.Pull - cli.ImagePull: %w", err)
	}

	_ = uc.imageRepo.DeleteByMachine(ctx, machineID)

	resp, err := cli.ImageInspect(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase.Pull - cli.ImageInspect: %w", err)
	}

	img := inspectToEntity(&resp, machineID)
	return &img, nil
}

// Delete removes a Docker image on the specified machine.
func (uc *ImageUseCase) Delete(ctx context.Context, machineID, imageID string) ([]dockerImage.DeleteResponse, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase.Delete - getDockerClient: %w", err)
	}
	defer cli.Close()

	resp, err := cli.ImageRemove(ctx, imageID, dockerImage.RemoveOptions{})
	if err != nil {
		if isImageInUseError(err) {
			return nil, fmt.Errorf(
				"ImageUseCase.Delete - cli.ImageRemove: %w",
				pkgErrors.Wrap(pkgErrors.ErrImageInUse, "stop or remove the running container that uses this image before deleting it"),
			)
		}

		return nil, fmt.Errorf("ImageUseCase.Delete - cli.ImageRemove: %w", err)
	}

	_ = uc.imageRepo.DeleteByMachine(ctx, machineID)

	return resp, nil
}

func isImageInUseError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "conflict: unable to delete") &&
		strings.Contains(msg, "image is being used by running container")
}

// Prune removes unused Docker images on the specified machine.
func (uc *ImageUseCase) Prune(ctx context.Context, machineID string) (*dockerImage.PruneReport, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase.Prune - getDockerClient: %w", err)
	}
	defer cli.Close()

	report, err := cli.ImagePrune(ctx, filters.NewArgs())
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase.Prune - cli.ImagePrune: %w", err)
	}

	_ = uc.imageRepo.DeleteByMachine(ctx, machineID)

	return &report, nil
}

// getDockerClient creates the appropriate Docker client for the given machine.
// For local machines, it connects to the local Docker socket.
// For remote machines, it connects via SSH tunnel.
func (uc *ImageUseCase) getDockerClient(ctx context.Context, machineID string) (dockerclient.Client, error) {
	if uc.testDockerClient != nil {
		return uc.testDockerClient, nil
	}

	m, err := uc.remoteMachineRepo.GetByID(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase.getDockerClient - remoteMachineRepo.GetByID: %w", err)
	}

	if m.ID == localMachineID {
		cli, err := dockerclient.NewLocalClient()
		if err != nil {
			return nil, fmt.Errorf("ImageUseCase.getDockerClient - dockerclient.NewLocalClient: %w", err)
		}
		return cli, nil
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase.getDockerClient - sshclient.New: %w", err)
	}

	cli, err := dockerclient.NewRemoteClient(sshClient, m.DockerHost)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("ImageUseCase.getDockerClient - dockerclient.NewRemoteClient: %w", err)
	}

	return cli, nil
}

// buildSSHConfig creates an SSH configuration from a remote machine entity.
func (uc *ImageUseCase) buildSSHConfig(m *entity.RemoteMachine) sshclient.Config {
	cfg := sshclient.Config{
		Host:    m.Host,
		Port:    m.Port,
		User:    m.Username,
		Timeout: 30 * time.Second,
	}

	switch m.AuthMethod {
	case entity.AuthMethodPassword:
		cfg.Password = m.Password
	case entity.AuthMethodKey:
		if m.SSHKey != "" {
			cfg.PrivateKey = []byte(m.SSHKey)
		} else if m.SSHKeyPath != "" {
			cfg.KeyPath = m.SSHKeyPath
		}
	}

	return cfg
}

// fetchImagesFromDocker fetches images from Docker and caches them.
func (uc *ImageUseCase) fetchImagesFromDocker(ctx context.Context, machineID string) ([]entity.Image, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase.fetchImagesFromDocker - getDockerClient: %w", err)
	}
	defer cli.Close()

	images, err := cli.ImageList(ctx, dockerImage.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("ImageUseCase.fetchImagesFromDocker - cli.ImageList: %w", err)
	}

	now := time.Now()
	for i := range images {
		images[i].MachineID = machineID
		images[i].CachedAt = now
	}

	if len(images) > 0 {
		if err := uc.imageRepo.UpsertBatch(ctx, machineID, images); err != nil {
			return images, fmt.Errorf("ImageUseCase.fetchImagesFromDocker - imageRepo.UpsertBatch: %w", err)
		}
	}

	return images, nil
}

// refresh runs in a goroutine to refresh the image cache.
func (uc *ImageUseCase) refresh(machineID string) {
	ctx := context.Background()

	_, err := uc.fetchImagesFromDocker(ctx, machineID)
	if err != nil {
		uc.l.Warn("ImageUseCase.refresh: %v", err)
	}
}

// inspectToEntity converts a Docker InspectResponse to an entity.Image.
func inspectToEntity(resp *dockerImage.InspectResponse, machineID string) entity.Image {
	id := resp.ID
	// Strip "sha256:" prefix if present
	if strings.HasPrefix(id, "sha256:") {
		id = strings.TrimPrefix(id, "sha256:")
	}
	// Take first 12 chars for short ID (consistent with ImageList conversion)
	if len(id) > 12 {
		id = id[:12]
	}

	// Parse Created timestamp (RFC 3339 nano-seconds)
	createdAt, err := time.Parse(time.RFC3339Nano, resp.Created)
	if err != nil {
		createdAt = time.Unix(0, 0)
	}

	var labels map[string]string
	if resp.Config != nil {
		labels = resp.Config.Labels
	}
	if labels == nil {
		labels = make(map[string]string)
	}

	repoTags := resp.RepoTags
	if repoTags == nil {
		repoTags = []string{}
	}

	repoDigests := resp.RepoDigests
	if repoDigests == nil {
		repoDigests = []string{}
	}

	return entity.Image{
		ID:          id,
		MachineID:   machineID,
		RepoTags:    repoTags,
		RepoDigests: repoDigests,
		Size:        resp.Size,
		CreatedAt:   createdAt,
		Labels:      labels,
	}
}
