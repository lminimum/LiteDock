CREATE TABLE IF NOT EXISTS volumes (
    name TEXT NOT NULL,
    machine_id TEXT NOT NULL,
    driver TEXT NOT NULL,
    mountpoint TEXT NOT NULL,
    created_at TEXT NOT NULL,
    scope TEXT NOT NULL,
    labels TEXT,
    size INTEGER NOT NULL DEFAULT 0,
    cached_at TIMESTAMP NOT NULL,
    PRIMARY KEY (machine_id, name)
);

CREATE INDEX idx_volumes_machine_id ON volumes(machine_id);
CREATE INDEX idx_volumes_cached_at ON volumes(cached_at);