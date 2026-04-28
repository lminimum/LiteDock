//go:build integration_test

package dockerclient

import (
	"context"
	"testing"
	"time"

	"github.com/lminimum/LiteDock/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestNewLocalClient_Success(t *testing.T) {
	cli, err := NewLocalClient()
	require.NoError(t, err)
	require.NotNil(t, cli)
	cli.Close()
}

func TestLocalClient_Ping(t *testing.T) {
	cli, err := NewLocalClient()
	require.NoError(t, err)
	defer cli.Close()

	err = cli.Ping(context.Background())
	require.NoError(t, err)
}

func TestLocalClient_ContainerList(t *testing.T) {
	cli, err := NewLocalClient()
	require.NoError(t, err)
	defer cli.Close()

	containers, err := cli.ContainerList(context.Background())
	require.NoError(t, err)
	require.NotNil(t, containers)

	for _, c := range containers {
		require.NotEmpty(t, c.ID, "container ID should not be empty")
		require.NotEmpty(t, c.Name, "container name should not be empty")
		require.NotEmpty(t, c.Image, "container image should not be empty")
		require.NotEmpty(t, c.Status, "container status should not be empty")
	}
}

func TestLocalClient_ContainerStart(t *testing.T) {
	cli, err := NewLocalClient()
	require.NoError(t, err)
	defer cli.Close()

	err = cli.ContainerStart(context.Background(), "nonexistent-test-container")
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrDockerOperation)
}

func TestLocalClient_ContainerStop(t *testing.T) {
	cli, err := NewLocalClient()
	require.NoError(t, err)
	defer cli.Close()

	err = cli.ContainerStop(context.Background(), "nonexistent-test-container", 5*time.Second)
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrDockerOperation)
}

func TestLocalClient_ContainerRestart(t *testing.T) {
	cli, err := NewLocalClient()
	require.NoError(t, err)
	defer cli.Close()

	err = cli.ContainerRestart(context.Background(), "nonexistent-test-container", 5*time.Second)
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrDockerOperation)
}

func TestLocalClient_ContainerRemove(t *testing.T) {
	cli, err := NewLocalClient()
	require.NoError(t, err)
	defer cli.Close()

	err = cli.ContainerRemove(context.Background(), "nonexistent-test-container", false)
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrDockerOperation)
}

func TestLocalClient_ContainerLogs(t *testing.T) {
	cli, err := NewLocalClient()
	require.NoError(t, err)
	defer cli.Close()

	_, err = cli.ContainerLogs(context.Background(), "nonexistent-test-container", "10")
	require.Error(t, err)
	require.ErrorIs(t, err, errors.ErrDockerOperation)
}
