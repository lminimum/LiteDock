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

type UseCase struct {
	repo           repo.RemoteMachineRepo
	containerRepo  repo.ContainerRepo
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

func (uc *UseCase) Update(ctx context.Context, m *entity.RemoteMachine) error {
	err := uc.repo.Update(ctx, m)
	if err != nil {
		return errors.Wrap(err, "UseCase.Update.repo.Update")
	}
	return nil
}

func (uc *UseCase) Delete(ctx context.Context, id string) error {
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
	m, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "UseCase.TestConnection.repo.GetByID")
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		return errors.Wrap(err, "UseCase.TestConnection.sshclient.New")
	}
	defer sshClient.Close()

	dockerClient, err := dockerclient.NewRemoteClient(sshClient, m.DockerHost)
	if err != nil {
		return errors.Wrap(err, "UseCase.TestConnection.dockerclient.NewRemoteClient")
	}
	defer dockerClient.Close()

	err = dockerClient.Ping(ctx)
	if err != nil {
		return errors.Wrap(err, "UseCase.TestConnection.dockerClient.Ping")
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
	m, err := uc.repo.GetByID(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.fetchAndCacheContainers.GetByID")
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.fetchAndCacheContainers.sshclient.New")
	}
	defer sshClient.Close()

	dockerClient, err := dockerclient.NewRemoteClient(sshClient, m.DockerHost)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.fetchAndCacheContainers.dockerclient.NewRemoteClient")
	}
	defer dockerClient.Close()

	containers, err := dockerClient.ContainerList(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.fetchAndCacheContainers.dockerClient.ContainerList")
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

	m, err := uc.repo.GetByID(ctx, machineID)
	if err != nil {
		uc.l.Warn("UseCase.refreshContainers.GetByID: %v", err)
		return
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		uc.l.Warn("UseCase.refreshContainers.sshclient.New: %v", err)
		return
	}
	defer sshClient.Close()

	dockerClient, err := dockerclient.NewRemoteClient(sshClient, m.DockerHost)
	if err != nil {
		uc.l.Warn("UseCase.refreshContainers.dockerclient.NewRemoteClient: %v", err)
		return
	}
	defer dockerClient.Close()

	containers, err := dockerClient.ContainerList(ctx)
	if err != nil {
		uc.l.Warn("UseCase.refreshContainers.dockerClient.ContainerList: %v", err)
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
	m, err := uc.repo.GetByID(ctx, machineID)
	if err != nil {
		return "", errors.Wrap(err, "UseCase.GetContainerLogs.repo.GetByID")
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		return "", errors.Wrap(err, "UseCase.GetContainerLogs.sshclient.New")
	}
	defer sshClient.Close()

	dockerClient, err := dockerclient.NewRemoteClient(sshClient, m.DockerHost)
	if err != nil {
		return "", errors.Wrap(err, "UseCase.GetContainerLogs.dockerclient.NewRemoteClient")
	}
	defer dockerClient.Close()

	logs, err := dockerClient.ContainerLogs(ctx, containerID, tail)
	if err != nil {
		return "", errors.Wrap(err, "UseCase.GetContainerLogs.dockerClient.ContainerLogs")
	}

	return logs, nil
}

func (uc *UseCase) ExecContainer(ctx context.Context, machineID, containerID string, cmd []string) (string, error) {
	m, err := uc.repo.GetByID(ctx, machineID)
	if err != nil {
		return "", errors.Wrap(err, "UseCase.ExecContainer.repo.GetByID")
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		return "", errors.Wrap(err, "UseCase.ExecContainer.sshclient.New")
	}
	defer sshClient.Close()

	dockerClient, err := dockerclient.NewRemoteClient(sshClient, m.DockerHost)
	if err != nil {
		return "", errors.Wrap(err, "UseCase.ExecContainer.dockerclient.NewRemoteClient")
	}
	defer dockerClient.Close()

	output, err := dockerClient.ContainerExec(ctx, containerID, cmd)
	if err != nil {
		return output, errors.Wrap(err, "UseCase.ExecContainer.dockerClient.ContainerExec")
	}

	return output, nil
}

func (uc *UseCase) StartContainer(ctx context.Context, machineID, containerID string) error {
	m, err := uc.repo.GetByID(ctx, machineID)
	if err != nil {
		return errors.Wrap(err, "UseCase.StartContainer.repo.GetByID")
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		return errors.Wrap(err, "UseCase.StartContainer.sshclient.New")
	}
	defer sshClient.Close()

	dockerClient, err := dockerclient.NewRemoteClient(sshClient, m.DockerHost)
	if err != nil {
		return errors.Wrap(err, "UseCase.StartContainer.dockerclient.NewRemoteClient")
	}
	defer dockerClient.Close()

	err = dockerClient.ContainerStart(ctx, containerID)
	if err != nil {
		return errors.Wrap(err, "UseCase.StartContainer.dockerClient.ContainerStart")
	}

	_ = uc.containerRepo.DeleteByMachine(ctx, machineID)
	return nil
}

func (uc *UseCase) StopContainer(ctx context.Context, machineID, containerID string) error {
	m, err := uc.repo.GetByID(ctx, machineID)
	if err != nil {
		return errors.Wrap(err, "UseCase.StopContainer.repo.GetByID")
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		return errors.Wrap(err, "UseCase.StopContainer.sshclient.New")
	}
	defer sshClient.Close()

	dockerClient, err := dockerclient.NewRemoteClient(sshClient, m.DockerHost)
	if err != nil {
		return errors.Wrap(err, "UseCase.StopContainer.dockerclient.NewRemoteClient")
	}
	defer dockerClient.Close()

	err = dockerClient.ContainerStop(ctx, containerID, 0)
	if err != nil {
		return errors.Wrap(err, "UseCase.StopContainer.dockerClient.ContainerStop")
	}

	_ = uc.containerRepo.DeleteByMachine(ctx, machineID)
	return nil
}

func (uc *UseCase) RestartContainer(ctx context.Context, machineID, containerID string) error {
	m, err := uc.repo.GetByID(ctx, machineID)
	if err != nil {
		return errors.Wrap(err, "UseCase.RestartContainer.repo.GetByID")
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		return errors.Wrap(err, "UseCase.RestartContainer.sshclient.New")
	}
	defer sshClient.Close()

	dockerClient, err := dockerclient.NewRemoteClient(sshClient, m.DockerHost)
	if err != nil {
		return errors.Wrap(err, "UseCase.RestartContainer.dockerclient.NewRemoteClient")
	}
	defer dockerClient.Close()

	err = dockerClient.ContainerRestart(ctx, containerID, 0)
	if err != nil {
		return errors.Wrap(err, "UseCase.RestartContainer.dockerClient.ContainerRestart")
	}

	_ = uc.containerRepo.DeleteByMachine(ctx, machineID)
	return nil
}

func (uc *UseCase) RemoveContainer(ctx context.Context, machineID, containerID string, force bool) error {
	m, err := uc.repo.GetByID(ctx, machineID)
	if err != nil {
		return errors.Wrap(err, "UseCase.RemoveContainer.repo.GetByID")
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		return errors.Wrap(err, "UseCase.RemoveContainer.sshclient.New")
	}
	defer sshClient.Close()

	dockerClient, err := dockerclient.NewRemoteClient(sshClient, m.DockerHost)
	if err != nil {
		return errors.Wrap(err, "UseCase.RemoveContainer.dockerclient.NewRemoteClient")
	}
	defer dockerClient.Close()

	err = dockerClient.ContainerRemove(ctx, containerID, force)
	if err != nil {
		return errors.Wrap(err, "UseCase.RemoveContainer.dockerClient.ContainerRemove")
	}

	_ = uc.containerRepo.DeleteByMachine(ctx, machineID)
	return nil
}

func (uc *UseCase) InspectContainer(ctx context.Context, machineID, containerID string) (interface{}, error) {
	m, err := uc.repo.GetByID(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.InspectContainer.repo.GetByID")
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.InspectContainer.sshclient.New")
	}
	defer sshClient.Close()

	dockerClient, err := dockerclient.NewRemoteClient(sshClient, m.DockerHost)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.InspectContainer.dockerclient.NewRemoteClient")
	}
	defer dockerClient.Close()

	c, err := dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.InspectContainer.dockerClient.ContainerInspect")
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
