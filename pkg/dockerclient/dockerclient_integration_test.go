//go:build integration_test
// +build integration_test

package dockerclient

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewRemoteClientFromConfig(t *testing.T) {
	cfg := SSHClientConfig{
		Host:     "127.0.0.1",
		Port:     22,
		User:     "testuser",
		Password: "testpass",
		Timeout:  5 * time.Second,
	}

	_, err := NewRemoteClientFromConfig(cfg, "/var/run/docker.sock")
	require.Error(t, err)
	require.Contains(t, err.Error(), "sshclient.New")
}

func TestSSHClientConfigAlias(t *testing.T) {
	cfg := SSHClientConfig{
		Host:       "localhost",
		Port:       22,
		User:       "root",
		PrivateKey: []byte("invalid-key"),
	}

	require.Equal(t, "localhost", cfg.Host)
	require.Equal(t, 22, cfg.Port)
	require.Equal(t, "root", cfg.User)
}

func TestDialerNilSSHClient(t *testing.T) {
	dialer := &sshDialer{
		sshClient: nil,
		sock:      "/var/run/docker.sock",
	}

	require.NotNil(t, dialer)
	require.Equal(t, "/var/run/docker.sock", dialer.sock)
	require.Nil(t, dialer.sshClient)
}

func TestHTTPClientTransport(t *testing.T) {
	cfg := SSHClientConfig{
		Host:     "localhost",
		Port:     22,
		User:     "test",
		Password: "test",
		Timeout:  5 * time.Second,
	}

	client, err := NewRemoteClientFromConfig(cfg, "/var/run/docker.sock")
	if err != nil {
		require.Contains(t, err.Error(), "sshclient.New")
		return
	}

	require.NotNil(t, client)
	require.NotNil(t, client.docker)
	require.NotNil(t, client.ssh)
	require.Equal(t, "/var/run/docker.sock", client.sock)

	client.Close()
}

func TestRemoteClientClose(t *testing.T) {
	rc := &RemoteClient{
		docker: nil,
		ssh:    nil,
		sock:   "",
	}

	err := rc.Close()
	require.NoError(t, err)

	rc = &RemoteClient{
		docker: nil,
		ssh:    nil,
		sock:   "/var/run/docker.sock",
	}

	err = rc.Close()
	require.NoError(t, err)
}
