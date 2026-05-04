package dockerclient

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/network"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestInspectToEntity_LocalClient(t *testing.T) {
	tests := []struct {
		name     string
		input    network.Inspect
		expected entity.Network
	}{
		{
			name: "full network",
			input: network.Inspect{
				Name:       "test-network",
				ID:         "abc123",
				Driver:     "bridge",
				Scope:      "local",
				Internal:   false,
				Attachable: false,
				EnableIPv6: false,
				Created:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Labels:     map[string]string{"env": "test"},
			},
			expected: entity.Network{
				ID:         "abc123",
				Name:       "test-network",
				Driver:     "bridge",
				Scope:      "local",
				Internal:   false,
				Attachable: false,
				EnableIPv6: false,
				Created:    "2024-01-15T10:30:00Z",
				Labels:     map[string]string{"env": "test"},
			},
		},
		{
			name: "overlay network",
			input: network.Inspect{
				Name:       "overlay-net",
				ID:         "def456",
				Driver:     "overlay",
				Scope:      "swarm",
				Internal:   true,
				Attachable: true,
				EnableIPv6: true,
				Created:    time.Date(2024, 6, 20, 15, 45, 30, 0, time.UTC),
				Labels:     nil,
			},
			expected: entity.Network{
				ID:         "def456",
				Name:       "overlay-net",
				Driver:     "overlay",
				Scope:      "swarm",
				Internal:   true,
				Attachable: true,
				EnableIPv6: true,
				Created:    "2024-06-20T15:45:30Z",
				Labels:     nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toEntity(tt.input)
			require.Equal(t, tt.expected.ID, result.ID)
			require.Equal(t, tt.expected.Name, result.Name)
			require.Equal(t, tt.expected.Driver, result.Driver)
			require.Equal(t, tt.expected.Scope, result.Scope)
			require.Equal(t, tt.expected.Internal, result.Internal)
			require.Equal(t, tt.expected.Attachable, result.Attachable)
			require.Equal(t, tt.expected.EnableIPv6, result.EnableIPv6)
			require.Equal(t, tt.expected.Created, result.Created)
			require.Equal(t, tt.expected.Labels, result.Labels)
		})
	}
}

// mockDockerClient implements the docker client methods for network testing
type mockDockerClient struct {
	networkList    func(ctx context.Context, options network.ListOptions) ([]network.Summary, error)
	networkCreate  func(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error)
	networkRemove  func(ctx context.Context, networkID string) error
	networkInspect func(ctx context.Context, networkID string, options network.InspectOptions) (network.Inspect, error)
}

func (m *mockDockerClient) NetworkList(ctx context.Context, options network.ListOptions) ([]network.Summary, error) {
	if m.networkList != nil {
		return m.networkList(ctx, options)
	}
	return nil, nil
}

func (m *mockDockerClient) NetworkCreate(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error) {
	if m.networkCreate != nil {
		return m.networkCreate(ctx, name, options)
	}
	return network.CreateResponse{}, nil
}

func (m *mockDockerClient) NetworkRemove(ctx context.Context, networkID string) error {
	if m.networkRemove != nil {
		return m.networkRemove(ctx, networkID)
	}
	return nil
}

func (m *mockDockerClient) NetworkInspect(ctx context.Context, networkID string, options network.InspectOptions) (network.Inspect, error) {
	if m.networkInspect != nil {
		return m.networkInspect(ctx, networkID, options)
	}
	return network.Inspect{}, nil
}

func TestLocalClientNetworkList(t *testing.T) {
	cli, err := NewLocalClient()
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer cli.Close()

	ctx := context.Background()
	networks, err := cli.NetworkList(ctx)
	require.NoError(t, err)
	// Just verify it doesn't error and returns a slice
	require.NotNil(t, networks)
}

func TestLocalClientNetworkInspect(t *testing.T) {
	cli, err := NewLocalClient()
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer cli.Close()

	ctx := context.Background()

	// First get a list of networks
	networks, err := cli.NetworkList(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, networks)

	// Inspect the first network
	networkID := networks[0].ID
	net, err := cli.NetworkInspect(ctx, networkID)
	require.NoError(t, err)
	require.NotNil(t, net)
	require.Equal(t, networkID, net.ID)
}

func TestLocalClientNetworkCreateDelete(t *testing.T) {
	cli, err := NewLocalClient()
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer cli.Close()

	ctx := context.Background()
	testNetworkName := "test-litedock-network-" + time.Now().Format("20060102150405")

	// Create a network
	net, err := cli.NetworkCreate(ctx, testNetworkName, "bridge")
	require.NoError(t, err)
	require.NotNil(t, net)
	require.NotEmpty(t, net.ID)
	require.Equal(t, testNetworkName, net.Name)
	require.Equal(t, "bridge", net.Driver)

	// Clean up - delete the network
	err = cli.NetworkDelete(ctx, net.ID)
	require.NoError(t, err)
}
