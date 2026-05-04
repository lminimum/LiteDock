package sshclient

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	_defaultTimeout    = 30 * time.Second
	_defaultPort       = 22
	_defaultTTL        = 5 * time.Minute
	_defaultMaxClients = 10
)

// Config holds SSH connection parameters.
type Config struct {
	Host       string
	Port       int
	User       string
	Password   string
	PrivateKey []byte
	KeyPath    string
	Timeout    time.Duration
}

// Client wraps an SSH client with connection management.
type Client struct {
	conn *ssh.Client
	cfg  Config
	mu   sync.RWMutex
}

// poolEntry tracks a client with its last access time.
type poolEntry struct {
	client     *Client
	lastAccess time.Time
}

// Pool manages SSH connections with TTL-based eviction.
type Pool struct {
	clients    map[string]*poolEntry
	mu         sync.RWMutex
	ttl        time.Duration
	maxClients int
	done       chan struct{}
}

// NewPool creates a new connection pool with TTL eviction.
func NewPool(ttl time.Duration, maxClients int) *Pool {
	if ttl <= 0 {
		ttl = _defaultTTL
	}

	if maxClients <= 0 {
		maxClients = _defaultMaxClients
	}

	p := &Pool{
		clients:    make(map[string]*poolEntry),
		ttl:        ttl,
		maxClients: maxClients,
		done:       make(chan struct{}),
	}

	go p.evictLoop()

	return p
}

// evictLoop periodically removes expired clients from the pool.
func (p *Pool) evictLoop() {
	ticker := time.NewTicker(p.ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.evict()
		case <-p.done:
			return
		}
	}
}

// evict removes clients that have exceeded their TTL.
func (p *Pool) evict() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()

	for key, entry := range p.clients {
		if now.Sub(entry.lastAccess) > p.ttl {
			_ = entry.client.Close()
			delete(p.clients, key)
		}
	}
}

// poolKey generates a unique key for a connection config.
func poolKey(cfg Config) string {
	port := cfg.Port
	if port == 0 {
		port = _defaultPort
	}

	return fmt.Sprintf("%s:%d:%s", cfg.Host, port, cfg.User)
}

// Get retrieves or creates an SSH client for the given config.
func (p *Pool) Get(ctx context.Context, cfg Config) (*Client, error) {
	key := poolKey(cfg)

	p.mu.RLock()
	entry, exists := p.clients[key]
	p.mu.RUnlock()

	if exists && entry.client.IsConnected() {
		p.mu.Lock()
		entry.lastAccess = time.Now()
		p.mu.Unlock()

		return entry.client, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock.
	if entry, exists = p.clients[key]; exists && entry.client.IsConnected() {
		entry.lastAccess = time.Now()
		return entry.client, nil
	}

	// Remove stale entry if exists.
	if exists {
		_ = entry.client.Close()
		delete(p.clients, key)
	}

	// Check pool capacity.
	if len(p.clients) >= p.maxClients {
		return nil, fmt.Errorf("sshclient.Pool.Get: pool capacity reached (%d)", p.maxClients)
	}

	client, err := New(cfg)
	if err != nil {
		return nil, fmt.Errorf("sshclient.Pool.Get: %w", err)
	}

	p.clients[key] = &poolEntry{
		client:     client,
		lastAccess: time.Now(),
	}

	return client, nil
}

// Close shuts down the pool and all managed connections.
func (p *Pool) Close() {
	close(p.done)

	p.mu.Lock()
	defer p.mu.Unlock()

	for key, entry := range p.clients {
		_ = entry.client.Close()
		delete(p.clients, key)
	}
}

// buildAuthMethods creates SSH auth methods from config.
func buildAuthMethods(cfg Config) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	if len(cfg.PrivateKey) > 0 {
		signer, err := ssh.ParsePrivateKey(cfg.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("sshclient.buildAuthMethods: parse private key: %w", err)
		}

		methods = append(methods, ssh.PublicKeys(signer))
	}

	if cfg.KeyPath != "" {
		keyData, err := os.ReadFile(cfg.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("sshclient.buildAuthMethods: read key file %q: %w", cfg.KeyPath, err)
		}

		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return nil, fmt.Errorf("sshclient.buildAuthMethods: parse key file: %w", err)
		}

		methods = append(methods, ssh.PublicKeys(signer))
	}

	if cfg.Password != "" {
		methods = append(methods, ssh.Password(cfg.Password))
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("sshclient.buildAuthMethods: no auth methods configured")
	}

	return methods, nil
}

// New creates a new SSH client and establishes a connection.
func New(cfg Config) (*Client, error) {
	if cfg.Port == 0 {
		cfg.Port = _defaultPort
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = _defaultTimeout
	}

	authMethods, err := buildAuthMethods(cfg)
	if err != nil {
		return nil, fmt.Errorf("sshclient.New: %w", err)
	}

	sshCfg := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // initial implementation
		Timeout:         cfg.Timeout,
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	conn, err := ssh.Dial("tcp", addr, sshCfg)
	if err != nil {
		return nil, fmt.Errorf("sshclient.New: dial %s: %w", addr, err)
	}

	return &Client{
		conn: conn,
		cfg:  cfg,
	}, nil
}

// Dial opens a network connection through the SSH tunnel.
func (c *Client) Dial(network, addr string) (net.Conn, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return nil, fmt.Errorf("sshclient.Client.Dial: connection closed")
	}

	conn, err := c.conn.Dial(network, addr)
	if err != nil {
		return nil, fmt.Errorf("sshclient.Client.Dial: %w", err)
	}

	return conn, nil
}

// Execute runs a command on the remote machine.
func (c *Client) Execute(ctx context.Context, cmd string) ([]byte, error) {
	c.mu.RLock()
	if c.conn == nil {
		c.mu.RUnlock()
		return nil, fmt.Errorf("sshclient.Client.Execute: connection closed")
	}
	c.mu.RUnlock()

	session, err := c.conn.NewSession()
	if err != nil {
		return nil, fmt.Errorf("sshclient.Client.Execute: new session: %w", err)
	}
	defer session.Close()

	// Handle context cancellation.
	done := make(chan struct{})

	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			_ = session.Signal(ssh.SIGKILL)
		case <-done:
		}
	}()

	var stdout bytes.Buffer

	session.Stdout = &stdout

	if err := session.Run(cmd); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("sshclient.Client.Execute: context canceled: %w", ctx.Err())
		}

		return nil, fmt.Errorf("sshclient.Client.Execute: run %q: %w", cmd, err)
	}

	return stdout.Bytes(), nil
}

// Close closes the SSH connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil

	if err != nil {
		return fmt.Errorf("sshclient.Client.Close: %w", err)
	}

	return nil
}

// IsConnected checks if the SSH connection is alive.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return false
	}

	// Send a keepalive request to verify the connection is alive.
	_, _, err := c.conn.SendRequest("keepalive@openssh.com", true, nil)

	return err == nil
}
