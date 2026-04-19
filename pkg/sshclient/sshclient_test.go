package sshclient

import (
	"testing"
	"time"
)

func TestConfigDefaults(t *testing.T) {
	cfg := Config{
		Host:     "localhost",
		Port:     22,
		User:     "test",
		Password: "test",
	}

	if cfg.Host != "localhost" {
		t.Errorf("expected host localhost, got %s", cfg.Host)
	}
	if cfg.Port != 22 {
		t.Errorf("expected port 22, got %d", cfg.Port)
	}
	if cfg.Timeout != 0 {
		t.Errorf("expected default timeout 0, got %v", cfg.Timeout)
	}
}

func TestConfigWithTimeout(t *testing.T) {
	cfg := Config{
		Host:     "localhost",
		Port:     22,
		User:     "test",
		Password: "test",
		Timeout:  30 * time.Second,
	}

	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", cfg.Timeout)
	}
}

func TestConfigKeyAuth(t *testing.T) {
	cfg := Config{
		Host:       "localhost",
		Port:       22,
		User:       "test",
		PrivateKey: []byte("test-key-content"),
	}

	if len(cfg.PrivateKey) == 0 {
		t.Error("expected private key to be set")
	}
}

func TestPoolConfig(t *testing.T) {
	pool := NewPool(5*time.Minute, 10)

	if pool.ttl != 5*time.Minute {
		t.Errorf("expected 5m TTL, got %v", pool.ttl)
	}
	if pool.maxClients != 10 {
		t.Errorf("expected max 10 clients, got %d", pool.maxClients)
	}
}
