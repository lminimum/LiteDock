// Package volume implements Docker volume management business logic.
package volume

import (
	"context"
	"time"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/repo"
	"github.com/lminimum/LiteDock/pkg/dockerclient"
	"github.com/lminimum/LiteDock/pkg/errors"
	"github.com/lminimum/LiteDock/pkg/logger"
	"github.com/lminimum/LiteDock/pkg/sshclient"
)

const (
	localMachineID   = "local"
	localMachineHost = "localhost"
)

// VolumeUseCase implements usecase.Volume for Docker volume management.
type VolumeUseCase struct {
	volumeRepo        repo.VolumeRepo
	remoteMachineRepo repo.RemoteMachineRepo
	cacheMaxAge       time.Duration
	l                 logger.Interface

	// testDockerClient is a test hook for injecting a mock dockerclient.Client.
	// It is nil in production and set only in tests.
	testDockerClient dockerclient.Client
}

// New creates a new VolumeUseCase.
func New(volumeRepo repo.VolumeRepo, rmRepo repo.RemoteMachineRepo, cacheMaxAge time.Duration, l logger.Interface) *VolumeUseCase {
	return &VolumeUseCase{
		volumeRepo:        volumeRepo,
		remoteMachineRepo: rmRepo,
		cacheMaxAge:       cacheMaxAge,
		l:                 l,
	}
}

// ListVolumes returns volumes for a machine with a cache-first strategy.
// On first call (empty cache), it fetches from Docker and caches the result.
// Subsequent calls return cached data while triggering an async refresh if the cache is stale.
func (uc *VolumeUseCase) ListVolumes(ctx context.Context, machineID string) ([]entity.Volume, error) {
	volumes, err := uc.volumeRepo.ListByMachine(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "VolumeUseCase.ListVolumes.ListByMachine")
	}

	valid, err := uc.volumeRepo.IsCacheValid(ctx, machineID, uc.cacheMaxAge)
	if err != nil {
		uc.l.Warn("VolumeUseCase.ListVolumes.IsCacheValid: %v", err)
	}

	if len(volumes) == 0 {
		return uc.fetchVolumesFromDocker(ctx, machineID)
	}

	if !valid {
		go uc.refreshVolumes(machineID)
	}

	return volumes, nil
}

// CreateVolume creates a Docker volume on the specified machine.
func (uc *VolumeUseCase) CreateVolume(ctx context.Context, machineID, name, driver string) (*entity.Volume, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "VolumeUseCase.CreateVolume.getDockerClient")
	}
	defer cli.Close()

	volume, err := cli.VolumeCreate(ctx, name, driver)
	if err != nil {
		return nil, errors.Wrap(err, "VolumeUseCase.CreateVolume.cli.VolumeCreate")
	}

	_ = uc.volumeRepo.DeleteByMachine(ctx, machineID)

	return volume, nil
}

// DeleteVolume deletes a Docker volume on the specified machine.
func (uc *VolumeUseCase) DeleteVolume(ctx context.Context, machineID, volumeName string) error {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return errors.Wrap(err, "VolumeUseCase.DeleteVolume.getDockerClient")
	}
	defer cli.Close()

	err = cli.VolumeDelete(ctx, volumeName)
	if err != nil {
		return errors.Wrap(err, "VolumeUseCase.DeleteVolume.cli.VolumeDelete")
	}

	_ = uc.volumeRepo.DeleteByMachine(ctx, machineID)

	return nil
}

// InspectVolume returns detailed information about a Docker volume.
func (uc *VolumeUseCase) InspectVolume(ctx context.Context, machineID, volumeName string) (*entity.Volume, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "VolumeUseCase.InspectVolume.getDockerClient")
	}
	defer cli.Close()

	volume, err := cli.VolumeInspect(ctx, volumeName)
	if err != nil {
		return nil, errors.Wrap(err, "VolumeUseCase.InspectVolume.cli.VolumeInspect")
	}

	return volume, nil
}

// getDockerClient creates the appropriate Docker client for the given machine.
// For local machines, it connects to the local Docker socket.
// For remote machines, it connects via SSH tunnel.
func (uc *VolumeUseCase) getDockerClient(ctx context.Context, machineID string) (dockerclient.Client, error) {
	if uc.testDockerClient != nil {
		return uc.testDockerClient, nil
	}

	m, err := uc.remoteMachineRepo.GetByID(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "VolumeUseCase.getDockerClient.GetByID")
	}

	if m.ID == localMachineID {
		cli, err := dockerclient.NewLocalClient()
		if err != nil {
			return nil, errors.Wrap(err, "VolumeUseCase.getDockerClient.NewLocalClient")
		}
		return cli, nil
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		return nil, errors.Wrap(err, "VolumeUseCase.getDockerClient.sshclient.New")
	}

	cli, err := dockerclient.NewRemoteClient(sshClient, m.DockerHost)
	if err != nil {
		sshClient.Close()
		return nil, errors.Wrap(err, "VolumeUseCase.getDockerClient.NewRemoteClient")
	}

	return cli, nil
}

// buildSSHConfig creates an SSH configuration from a remote machine entity.
func (uc *VolumeUseCase) buildSSHConfig(m *entity.RemoteMachine) sshclient.Config {
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

// fetchVolumesFromDocker fetches volumes from Docker and caches them.
func (uc *VolumeUseCase) fetchVolumesFromDocker(ctx context.Context, machineID string) ([]entity.Volume, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "VolumeUseCase.fetchVolumesFromDocker.getDockerClient")
	}
	defer cli.Close()

	volumes, err := cli.VolumeList(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "VolumeUseCase.fetchVolumesFromDocker.cli.VolumeList")
	}

	now := time.Now()
	for i := range volumes {
		volumes[i].MachineID = machineID
		volumes[i].CachedAt = now
	}

	if len(volumes) > 0 {
		if err := uc.volumeRepo.UpsertBatch(ctx, machineID, volumes); err != nil {
			return volumes, errors.Wrap(err, "VolumeUseCase.fetchVolumesFromDocker.UpsertBatch")
		}
	}

	return volumes, nil
}

// refreshVolumes runs in a goroutine to refresh the volume cache.
func (uc *VolumeUseCase) refreshVolumes(machineID string) {
	ctx := context.Background()

	_, err := uc.fetchVolumesFromDocker(ctx, machineID)
	if err != nil {
		uc.l.Warn("VolumeUseCase.refreshVolumes: %v", err)
	}
}
