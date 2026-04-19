-- Create remote_machines table
CREATE TABLE IF NOT EXISTS remote_machines (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER DEFAULT 22,
    username VARCHAR(255) NOT NULL,
    auth_method VARCHAR(50) DEFAULT 'password',
    password VARCHAR(255),
    ssh_key_path VARCHAR(512),
    ssh_key TEXT,
    docker_host VARCHAR(512) DEFAULT '/var/run/docker.sock',
    status VARCHAR(50) DEFAULT 'unknown',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on name for faster lookups
CREATE INDEX idx_remote_machines_name ON remote_machines(name);

-- Create index on host for faster lookups
CREATE INDEX idx_remote_machines_host ON remote_machines(host);
