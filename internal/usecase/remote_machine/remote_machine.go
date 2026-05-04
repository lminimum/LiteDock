package remote_machine

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/repo"
	"github.com/lminimum/LiteDock/pkg/dockerclient"
	"github.com/lminimum/LiteDock/pkg/errors"
	"github.com/lminimum/LiteDock/pkg/logger"
	"github.com/lminimum/LiteDock/pkg/sshclient"
)

// LocalMachineHost is the host value used to identify a local Docker machine.
const LocalMachineHost = "localhost"

// LocalMachineID is the fixed ID for the built-in local machine.
const LocalMachineID = "local"

type UseCase struct {
	repo          repo.RemoteMachineRepo
	containerRepo repo.ContainerRepo
	cacheMaxAge   time.Duration
	l             logger.Interface
}

func New(repo repo.RemoteMachineRepo, containerRepo repo.ContainerRepo, cacheMaxAge time.Duration, l logger.Interface) *UseCase {
	return &UseCase{
		repo:          repo,
		containerRepo: containerRepo,
		cacheMaxAge:   cacheMaxAge,
		l:             l,
	}
}

// isLocalMachine returns true if the machine represents the local Docker socket.
func isLocalMachine(m *entity.RemoteMachine) bool {
	return m.ID == LocalMachineID
}

// getDockerClient creates the appropriate Docker client for the given machine.
// For local machines, it connects directly to the local Docker socket.
// For remote machines, it connects via SSH tunnel.
func (uc *UseCase) getDockerClient(ctx context.Context, machineID string) (dockerclient.Client, error) {
	m, err := uc.repo.GetByID(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.getDockerClient.GetByID")
	}

	if isLocalMachine(m) {
		cli, err := dockerclient.NewLocalClient()
		if err != nil {
			return nil, errors.Wrap(err, "UseCase.getDockerClient.NewLocalClient")
		}
		uc.l.Debug("UseCase.getDockerClient: using local Docker socket for machine %s", machineID)
		return cli, nil
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.getDockerClient.sshclient.New")
	}

	cli, err := dockerclient.NewRemoteClient(sshClient, m.DockerHost)
	if err != nil {
		sshClient.Close()
		return nil, errors.Wrap(err, "UseCase.getDockerClient.NewRemoteClient")
	}

	return cli, nil
}

func (uc *UseCase) Create(ctx context.Context, m *entity.RemoteMachine) (*entity.RemoteMachine, error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	if m.Port == 0 {
		m.Port = 22
	}
	if m.DockerHost == "" {
		m.DockerHost = "/var/run/docker.sock"
	}
	if m.Status == "" {
		m.Status = "unknown"
	}

	err := uc.repo.Create(ctx, m)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.Create.repo.Create")
	}

	return m, nil
}

func (uc *UseCase) GetByID(ctx context.Context, id string) (*entity.RemoteMachine, error) {
	m, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.GetByID.repo.GetByID")
	}
	return m, nil
}

func (uc *UseCase) List(ctx context.Context) ([]entity.RemoteMachine, error) {
	machines, err := uc.repo.List(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.List.repo.List")
	}
	return machines, nil
}

func (uc *UseCase) Count(ctx context.Context) (int64, error) {
	count, err := uc.repo.Count(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "UseCase.Count.repo.Count")
	}
	return count, nil
}

func (uc *UseCase) Update(ctx context.Context, m *entity.RemoteMachine) error {
	if m.ID == LocalMachineID {
		return errors.Wrap(errors.ErrInvalidInput, "UseCase.Update: cannot update local machine")
	}
	err := uc.repo.Update(ctx, m)
	if err != nil {
		return errors.Wrap(err, "UseCase.Update.repo.Update")
	}
	return nil
}

func (uc *UseCase) Delete(ctx context.Context, id string) error {
	if id == LocalMachineID {
		return errors.Wrap(errors.ErrInvalidInput, "UseCase.Delete: cannot delete local machine")
	}
	err := uc.repo.Delete(ctx, id)
	if err != nil {
		return errors.Wrap(err, "UseCase.Delete.repo.Delete")
	}
	return nil
}

func (uc *UseCase) GetByHost(ctx context.Context, host string) (*entity.RemoteMachine, error) {
	m, err := uc.repo.GetByHost(ctx, host)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.GetByHost.repo.GetByHost")
	}
	return m, nil
}

func (uc *UseCase) TestConnection(ctx context.Context, id string) error {
	cli, err := uc.getDockerClient(ctx, id)
	if err != nil {
		return errors.Wrap(err, "UseCase.TestConnection.getDockerClient")
	}
	defer cli.Close()

	err = cli.Ping(ctx)
	if err != nil {
		return errors.Wrap(err, "UseCase.TestConnection.cli.Ping")
	}

	return nil
}

func (uc *UseCase) ListContainers(ctx context.Context, machineID string) ([]entity.Container, error) {
	containers, err := uc.containerRepo.ListByMachine(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.ListContainers.ListByMachine")
	}

	valid, err := uc.containerRepo.IsCacheValid(ctx, machineID, uc.cacheMaxAge)
	if err != nil {
		uc.l.Warn("UseCase.ListContainers.IsCacheValid: %v", err)
	}

	if len(containers) == 0 {
		return uc.fetchAndCacheContainers(ctx, machineID)
	}

	if !valid {
		go uc.refreshContainers(machineID)
	}

	return containers, nil
}

func (uc *UseCase) fetchAndCacheContainers(ctx context.Context, machineID string) ([]entity.Container, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.fetchAndCacheContainers.getDockerClient")
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.fetchAndCacheContainers.cli.ContainerList")
	}

	now := time.Now()
	for i := range containers {
		containers[i].MachineID = machineID
		containers[i].CachedAt = now
	}

	if len(containers) > 0 {
		err = uc.containerRepo.UpsertBatch(ctx, machineID, containers)
		if err != nil {
			uc.l.Warn("UseCase.fetchAndCacheContainers.UpsertBatch: %v", err)
		}
	}

	return containers, nil
}

func (uc *UseCase) refreshContainers(machineID string) {
	ctx := context.Background()

	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		uc.l.Warn("UseCase.refreshContainers.getDockerClient: %v", err)
		return
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx)
	if err != nil {
		uc.l.Warn("UseCase.refreshContainers.cli.ContainerList: %v", err)
		return
	}

	now := time.Now()
	for i := range containers {
		containers[i].MachineID = machineID
		containers[i].CachedAt = now
	}

	if len(containers) > 0 {
		err = uc.containerRepo.UpsertBatch(ctx, machineID, containers)
		if err != nil {
			uc.l.Warn("UseCase.refreshContainers.UpsertBatch: %v", err)
			return
		}
	}

	uc.l.Debug("UseCase.refreshContainers: refreshed %d containers for machine %s", len(containers), machineID)
}

func (uc *UseCase) GetContainerLogs(ctx context.Context, machineID, containerID, tail string) (string, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return "", errors.Wrap(err, "UseCase.GetContainerLogs.getDockerClient")
	}
	defer cli.Close()

	logs, err := cli.ContainerLogs(ctx, containerID, tail)
	if err != nil {
		return "", errors.Wrap(err, "UseCase.GetContainerLogs.cli.ContainerLogs")
	}

	return logs, nil
}

func (uc *UseCase) ExecContainer(ctx context.Context, machineID, containerID string, cmd []string) (string, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return "", errors.Wrap(err, "UseCase.ExecContainer.getDockerClient")
	}
	defer cli.Close()

	output, err := cli.ContainerExec(ctx, containerID, cmd)
	if err != nil {
		return output, errors.Wrap(err, "UseCase.ExecContainer.cli.ContainerExec")
	}

	return output, nil
}

func (uc *UseCase) StartContainer(ctx context.Context, machineID, containerID string) error {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return errors.Wrap(err, "UseCase.StartContainer.getDockerClient")
	}
	defer cli.Close()

	err = cli.ContainerStart(ctx, containerID)
	if err != nil {
		return errors.Wrap(err, "UseCase.StartContainer.cli.ContainerStart")
	}

	_ = uc.containerRepo.DeleteByMachine(ctx, machineID)
	return nil
}

func (uc *UseCase) StopContainer(ctx context.Context, machineID, containerID string) error {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return errors.Wrap(err, "UseCase.StopContainer.getDockerClient")
	}
	defer cli.Close()

	err = cli.ContainerStop(ctx, containerID, 0)
	if err != nil {
		return errors.Wrap(err, "UseCase.StopContainer.cli.ContainerStop")
	}

	_ = uc.containerRepo.DeleteByMachine(ctx, machineID)
	return nil
}

func (uc *UseCase) RestartContainer(ctx context.Context, machineID, containerID string) error {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return errors.Wrap(err, "UseCase.RestartContainer.getDockerClient")
	}
	defer cli.Close()

	err = cli.ContainerRestart(ctx, containerID, 0)
	if err != nil {
		return errors.Wrap(err, "UseCase.RestartContainer.cli.ContainerRestart")
	}

	_ = uc.containerRepo.DeleteByMachine(ctx, machineID)
	return nil
}

func (uc *UseCase) RemoveContainer(ctx context.Context, machineID, containerID string, force bool) error {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return errors.Wrap(err, "UseCase.RemoveContainer.getDockerClient")
	}
	defer cli.Close()

	err = cli.ContainerRemove(ctx, containerID, force)
	if err != nil {
		return errors.Wrap(err, "UseCase.RemoveContainer.cli.ContainerRemove")
	}

	_ = uc.containerRepo.DeleteByMachine(ctx, machineID)
	return nil
}

func (uc *UseCase) InspectContainer(ctx context.Context, machineID, containerID string) (interface{}, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.InspectContainer.getDockerClient")
	}
	defer cli.Close()

	c, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.InspectContainer.cli.ContainerInspect")
	}

	return c, nil
}

func (uc *UseCase) buildSSHConfig(m *entity.RemoteMachine) sshclient.Config {
	cfg := sshclient.Config{
		Host:    m.Host,
		Port:    m.Port,
		User:    m.Username,
		Timeout: 30 * time.Second,
	}

	if m.AuthMethod == entity.AuthMethodPassword {
		cfg.Password = m.Password
	} else if m.AuthMethod == entity.AuthMethodKey {
		if m.SSHKey != "" {
			cfg.PrivateKey = []byte(m.SSHKey)
		} else if m.SSHKeyPath != "" {
			cfg.KeyPath = m.SSHKeyPath
		}
	}

	return cfg
}
