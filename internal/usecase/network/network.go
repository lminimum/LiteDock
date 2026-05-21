// Package network implements Docker network management business logic.
package network

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

// builtInNetworks are Docker networks that cannot be deleted.
var builtInNetworks = map[string]bool{
	"bridge": true,
	"host":   true,
	"none":   true,
}

// NetworkUseCase implements usecase.Network for Docker network management.
type NetworkUseCase struct {
	networkRepo       repo.NetworkRepo
	remoteMachineRepo repo.RemoteMachineRepo
	cacheMaxAge       time.Duration
	l                 logger.Interface

	// testDockerClient is a test hook for injecting a mock dockerclient.Client.
	// It is nil in production and set only in tests.
	testDockerClient dockerclient.Client
}

// New creates a new NetworkUseCase.
func New(networkRepo repo.NetworkRepo, rmRepo repo.RemoteMachineRepo, cacheMaxAge time.Duration, l logger.Interface) *NetworkUseCase {
	return &NetworkUseCase{
		networkRepo:       networkRepo,
		remoteMachineRepo: rmRepo,
		cacheMaxAge:       cacheMaxAge,
		l:                 l,
	}
}

// ListNetworks returns networks for a machine with a cache-first strategy.
// On first call (empty cache), it fetches from Docker and caches the result.
// Subsequent calls return cached data while triggering an async refresh if the cache is stale.
func (uc *NetworkUseCase) ListNetworks(ctx context.Context, machineID string) ([]entity.Network, error) {
	networks, err := uc.networkRepo.ListByMachine(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "NetworkUseCase.ListNetworks.ListByMachine")
	}

	valid, err := uc.networkRepo.IsCacheValid(ctx, machineID, uc.cacheMaxAge)
	if err != nil {
		uc.l.Warn("NetworkUseCase.ListNetworks.IsCacheValid: %v", err)
	}

	if len(networks) == 0 {
		return uc.fetchNetworksFromDocker(ctx, machineID)
	}

	if !valid {
		go uc.refreshNetworks(machineID)
	}

	return networks, nil
}

// CreateNetwork creates a Docker network on the specified machine.
func (uc *NetworkUseCase) CreateNetwork(ctx context.Context, machineID, name, driver string) (*entity.Network, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "NetworkUseCase.CreateNetwork.getDockerClient")
	}
	defer cli.Close()

	network, err := cli.NetworkCreate(ctx, name, driver)
	if err != nil {
		return nil, errors.Wrap(err, "NetworkUseCase.CreateNetwork.cli.NetworkCreate")
	}

	_ = uc.networkRepo.DeleteByMachine(ctx, machineID)

	return network, nil
}

// DeleteNetwork deletes a Docker network on the specified machine.
// Built-in networks (bridge, host, none) cannot be deleted.
func (uc *NetworkUseCase) DeleteNetwork(ctx context.Context, machineID, networkName string) error {
	if builtInNetworks[networkName] {
		return errors.Wrap(errors.ErrInvalidInput, "NetworkUseCase.DeleteNetwork: cannot delete built-in network "+networkName)
	}

	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return errors.Wrap(err, "NetworkUseCase.DeleteNetwork.getDockerClient")
	}
	defer cli.Close()

	err = cli.NetworkDelete(ctx, networkName)
	if err != nil {
		return errors.Wrap(err, "NetworkUseCase.DeleteNetwork.cli.NetworkDelete")
	}

	_ = uc.networkRepo.DeleteByMachine(ctx, machineID)

	return nil
}

// InspectNetwork returns detailed information about a Docker network.
func (uc *NetworkUseCase) InspectNetwork(ctx context.Context, machineID, networkName string) (*entity.Network, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "NetworkUseCase.InspectNetwork.getDockerClient")
	}
	defer cli.Close()

	network, err := cli.NetworkInspect(ctx, networkName)
	if err != nil {
		return nil, errors.Wrap(err, "NetworkUseCase.InspectNetwork.cli.NetworkInspect")
	}

	return network, nil
}

// getDockerClient creates the appropriate Docker client for the given machine.
// For local machines, it connects to the local Docker socket.
// For remote machines, it connects via SSH tunnel.
func (uc *NetworkUseCase) getDockerClient(ctx context.Context, machineID string) (dockerclient.Client, error) {
	if uc.testDockerClient != nil {
		return uc.testDockerClient, nil
	}

	if machineID == localMachineID {
		cli, err := dockerclient.NewLocalClient()
		if err != nil {
			return nil, errors.Wrap(err, "NetworkUseCase.getDockerClient.NewLocalClient")
		}
		return cli, nil
	}

	m, err := uc.remoteMachineRepo.GetByID(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "NetworkUseCase.getDockerClient.GetByID")
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		return nil, errors.Wrap(err, "NetworkUseCase.getDockerClient.sshclient.New")
	}

	cli, err := dockerclient.NewRemoteClient(sshClient, m.DockerHost)
	if err != nil {
		sshClient.Close()
		return nil, errors.Wrap(err, "NetworkUseCase.getDockerClient.NewRemoteClient")
	}

	return cli, nil
}

// buildSSHConfig creates an SSH configuration from a remote machine entity.
func (uc *NetworkUseCase) buildSSHConfig(m *entity.RemoteMachine) sshclient.Config {
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

// fetchNetworksFromDocker fetches networks from Docker and caches them.
func (uc *NetworkUseCase) fetchNetworksFromDocker(ctx context.Context, machineID string) ([]entity.Network, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "NetworkUseCase.fetchNetworksFromDocker.getDockerClient")
	}
	defer cli.Close()

	networks, err := cli.NetworkList(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "NetworkUseCase.fetchNetworksFromDocker.cli.NetworkList")
	}

	now := time.Now()
	for i := range networks {
		networks[i].MachineID = machineID
		networks[i].CachedAt = now
	}

	if len(networks) > 0 {
		if err := uc.networkRepo.UpsertBatch(ctx, machineID, networks); err != nil {
			return networks, errors.Wrap(err, "NetworkUseCase.fetchNetworksFromDocker.UpsertBatch")
		}
	}

	return networks, nil
}

// refreshNetworks runs in a goroutine to refresh the network cache.
func (uc *NetworkUseCase) refreshNetworks(machineID string) {
	ctx := context.Background()

	_, err := uc.fetchNetworksFromDocker(ctx, machineID)
	if err != nil {
		uc.l.Warn("NetworkUseCase.refreshNetworks: %v", err)
	}
}
