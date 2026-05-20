package container

import (
	"context"
	"fmt"
	"time"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/repo"
	"github.com/lminimum/LiteDock/pkg/dockerclient"
	"github.com/lminimum/LiteDock/pkg/logger"
	"github.com/lminimum/LiteDock/pkg/sshclient"
)

const (
	localMachineID   = "local"
	localMachineHost = "localhost"
)

// UseCase handles container business logic.
type UseCase struct {
	repo interface {
		List(ctx context.Context) ([]entity.Container, error)
		Get(ctx context.Context, id string) (*entity.Container, error)
		CountAll(ctx context.Context) (int64, error)
		CountByStatus(ctx context.Context, status string) (int64, error)
	}
	remoteMachineRepo repo.RemoteMachineRepo
	l                 logger.Interface

	// testDockerClient is a test hook for injecting a mock dockerclient.Client.
	// It is nil in production and set only in tests.
	testDockerClient dockerclient.Client
}

// New creates a new container use case.
func New(
	repo interface {
		List(ctx context.Context) ([]entity.Container, error)
		Get(ctx context.Context, id string) (*entity.Container, error)
		CountAll(ctx context.Context) (int64, error)
		CountByStatus(ctx context.Context, status string) (int64, error)
	},
	remoteMachineRepo repo.RemoteMachineRepo,
	l logger.Interface,
) *UseCase {
	return &UseCase{repo: repo, remoteMachineRepo: remoteMachineRepo, l: l}
}

// List returns all containers.
func (uc *UseCase) List(ctx context.Context) ([]entity.Container, error) {
	return uc.repo.List(ctx)
}

// Get returns a container by ID.
func (uc *UseCase) Get(ctx context.Context, id string) (*entity.Container, error) {
	return uc.repo.Get(ctx, id)
}

// CountAll returns total container count.
func (uc *UseCase) CountAll(ctx context.Context) (int64, error) {
	return uc.repo.CountAll(ctx)
}

// CountByStatus returns container count by status.
func (uc *UseCase) CountByStatus(ctx context.Context, status string) (int64, error) {
	return uc.repo.CountByStatus(ctx, status)
}

// Start starts a container on the specified machine.
func (uc *UseCase) Start(ctx context.Context, machineID, containerID string) error {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return fmt.Errorf("UseCase.Start - getDockerClient: %w", err)
	}
	defer cli.Close()

	if err := cli.ContainerStart(ctx, containerID); err != nil {
		return fmt.Errorf("UseCase.Start - cli.ContainerStart: %w", err)
	}
	return nil
}

// Stop stops a container on the specified machine.
func (uc *UseCase) Stop(ctx context.Context, machineID, containerID string) error {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return fmt.Errorf("UseCase.Stop - getDockerClient: %w", err)
	}
	defer cli.Close()

	if err := cli.ContainerStop(ctx, containerID, 0); err != nil {
		return fmt.Errorf("UseCase.Stop - cli.ContainerStop: %w", err)
	}
	return nil
}

// Restart restarts a container on the specified machine.
func (uc *UseCase) Restart(ctx context.Context, machineID, containerID string) error {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return fmt.Errorf("UseCase.Restart - getDockerClient: %w", err)
	}
	defer cli.Close()

	if err := cli.ContainerRestart(ctx, containerID, 0); err != nil {
		return fmt.Errorf("UseCase.Restart - cli.ContainerRestart: %w", err)
	}
	return nil
}

// Logs returns recent log lines for a container.
func (uc *UseCase) Logs(ctx context.Context, machineID, containerID, tail string) (string, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return "", fmt.Errorf("UseCase.Logs - getDockerClient: %w", err)
	}
	defer cli.Close()

	logs, err := cli.ContainerLogs(ctx, containerID, tail)
	if err != nil {
		return "", fmt.Errorf("UseCase.Logs - cli.ContainerLogs: %w", err)
	}
	return logs, nil
}

// getDockerClient creates the appropriate Docker client for the given machine.
// For local machines, it connects to the local Docker socket.
// For remote machines, it connects via SSH tunnel.
func (uc *UseCase) getDockerClient(ctx context.Context, machineID string) (dockerclient.Client, error) {
	if uc.testDockerClient != nil {
		return uc.testDockerClient, nil
	}

	m, err := uc.remoteMachineRepo.GetByID(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("UseCase.getDockerClient - remoteMachineRepo.GetByID: %w", err)
	}

	if m.ID == localMachineID {
		cli, err := dockerclient.NewLocalClient()
		if err != nil {
			return nil, fmt.Errorf("UseCase.getDockerClient - dockerclient.NewLocalClient: %w", err)
		}
		return cli, nil
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		return nil, fmt.Errorf("UseCase.getDockerClient - sshclient.New: %w", err)
	}

	cli, err := dockerclient.NewRemoteClient(sshClient, m.DockerHost)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("UseCase.getDockerClient - dockerclient.NewRemoteClient: %w", err)
	}

	return cli, nil
}

// buildSSHConfig creates an SSH configuration from a remote machine entity.
func (uc *UseCase) buildSSHConfig(m *entity.RemoteMachine) sshclient.Config {
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
