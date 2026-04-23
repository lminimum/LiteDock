CREATE TABLE IF NOT EXISTS containers (
    id VARCHAR(255) PRIMARY KEY,
    machine_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    image VARCHAR(512) NOT NULL,
    status VARCHAR(50) NOT NULL,
    ports TEXT,
    created_at BIGINT NOT NULL,
    cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (machine_id) REFERENCES remote_machines(id) ON DELETE CASCADE
);

CREATE INDEX idx_containers_machine_id ON containers(machine_id);
CREATE INDEX idx_containers_cached_at ON containers(cached_at);
