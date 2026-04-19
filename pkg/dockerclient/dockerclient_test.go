package dockerclient

import (
	"testing"
	"time"

	"github.com/lminimum/LiteDock/pkg/sshclient"
)

func TestNewRemoteClientFromConfigInvalidSock(t *testing.T) {
	cfg := sshclient.Config{
		Host:     "localhost",
		Port:     22,
		User:     "test",
		Password: "test",
	}

	_, err := NewRemoteClientFromConfig(cfg, "/nonexistent/docker.sock")
	if err == nil {
		t.Error("expected error for invalid sock path")
	}
}

func TestSSHClientConfig(t *testing.T) {
	cfg := sshclient.Config{
		Host:     "localhost",
		Port:     22,
		User:     "testuser",
		Password: "testpass",
		Timeout:  30 * time.Second,
	}

	if cfg.Host != "localhost" {
		t.Errorf("expected host localhost, got %s", cfg.Host)
	}
	if cfg.Port != 22 {
		t.Errorf("expected port 22, got %d", cfg.Port)
	}
	if cfg.User != "testuser" {
		t.Errorf("expected user testuser, got %s", cfg.User)
	}
	if cfg.Password != "testpass" {
		t.Errorf("expected password testpass, got %s", cfg.Password)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", cfg.Timeout)
	}
}

func TestSSHClientConfigWithKey(t *testing.T) {
	cfg := sshclient.Config{
		Host:       "localhost",
		Port:       22,
		User:       "testuser",
		PrivateKey: []byte("test-key"),
		Timeout:    30 * time.Second,
	}

	if len(cfg.PrivateKey) == 0 {
		t.Error("expected private key to be set")
	}
	if cfg.Password != "" {
		t.Error("expected no password when using key auth")
	}
}
