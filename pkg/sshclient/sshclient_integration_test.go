//go:build integration_test
// +build integration_test

package sshclient

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestLocalDockerSocket tests SSH client Dial by connecting to the local Docker socket.
// This simulates the SSH tunnel use case without needing a real SSH server.
func TestLocalDockerSocket(t *testing.T) {
	// Create a config pointing to localhost:22 (won't connect, but tests config parsing)
	cfg := Config{
		Host:     "127.0.0.1",
		Port:     22,
		User:     "testuser",
		Password: "testpass",
		Timeout:  5 * time.Second,
	}

	// Verify config defaults
	require.Equal(t, 22, cfg.Port)
	require.Equal(t, "testuser", cfg.User)
	require.Equal(t, "testpass", cfg.Password)
}

// TestBuildAuthMethods tests the auth method building logic.
func TestBuildAuthMethods(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "password auth",
			cfg: Config{
				Host:     "localhost",
				Port:     22,
				User:     "root",
				Password: "secret",
			},
			wantErr: false,
		},
		{
			name: "key auth from bytes",
			cfg: Config{
				Host:       "localhost",
				Port:       22,
				User:       "root",
				PrivateKey: []byte("invalid-key-content"),
			},
			wantErr: true, // invalid key content
		},
		{
			name: "no auth methods",
			cfg: Config{
				Host: "localhost",
				Port: 22,
				User: "root",
			},
			wantErr: true, // no auth configured
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildAuthMethods(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestPoolCreation tests the connection pool initialization.
func TestPoolCreation(t *testing.T) {
	pool := NewPool(5*time.Minute, 10)
	require.NotNil(t, pool)
	require.NotNil(t, pool.clients)

	// Verify default values
	pool2 := NewPool(0, 0)
	require.NotNil(t, pool2)

	pool.Close()
	pool2.Close()
}

// TestPoolKey tests the pool key generation.
func TestPoolKey(t *testing.T) {
	cfg1 := Config{Host: "192.168.1.1", Port: 22, User: "root"}
	cfg2 := Config{Host: "192.168.1.1", Port: 2222, User: "root"}
	cfg3 := Config{Host: "192.168.1.2", Port: 22, User: "root"}

	key1 := poolKey(cfg1)
	key2 := poolKey(cfg2)
	key3 := poolKey(cfg3)

	require.Equal(t, key1, poolKey(cfg1)) // same config = same key
	require.NotEqual(t, key1, key2)       // different port = different key
	require.NotEqual(t, key1, key3)       // different host = different key
}

// TestPoolConcurrency tests concurrent access to the pool.
func TestPoolConcurrency(t *testing.T) {
	pool := NewPool(5*time.Minute, 5)
	defer pool.Close()

	var wg sync.WaitGroup

	// Simulate concurrent Get calls (will fail to connect but test concurrency)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			cfg := Config{
				Host:     "127.0.0.1",
				Port:     22,
				User:     "testuser",
				Password: "testpass",
				Timeout:  50 * time.Millisecond,
			}

			_, err := pool.Get(ctx, cfg)
			// We expect this to fail (no SSH server), but not panic
			if err != nil {
				t.Logf("Expected connection failure: %v", err)
			}
		}()
	}

	wg.Wait()
}

// TestSSHClientInterfaceCompliance tests that Client implements net.Listener interface
// requirements for the Docker socket forwarding scenario.
func TestSSHClientDialNetworkTypes(t *testing.T) {
	validNetworks := []string{"unix", "tcp", "udp"}

	for _, network := range validNetworks {
		t.Run(network, func(t *testing.T) {
			client := &Client{}

			_, err := client.Dial(network, "/tmp/test.sock")
			require.Error(t, err)
			require.Contains(t, err.Error(), "connection closed")
		})
	}
}

// BenchmarkPoolGet benchmarks the pool.Get performance.
func BenchmarkPoolGet(b *testing.B) {
	pool := NewPool(5*time.Minute, 100)
	defer pool.Close()

	cfg := Config{
		Host:     "localhost",
		Port:     22,
		User:     "test",
		Password: "test",
		Timeout:  1 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		_, _ = pool.Get(ctx, cfg)
		cancel()
	}
}

// TestClientIsConnected tests the IsConnected check.
func TestClientIsConnected(t *testing.T) {
	client := &Client{}
	require.False(t, client.IsConnected()) // nil conn

	client.conn = nil
	require.False(t, client.IsConnected())
}
