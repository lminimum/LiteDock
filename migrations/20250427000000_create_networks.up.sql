CREATE TABLE IF NOT EXISTS networks (
    id TEXT NOT NULL,
    machine_id TEXT NOT NULL,
    name TEXT NOT NULL,
    driver TEXT NOT NULL,
    scope TEXT NOT NULL,
    internal INTEGER NOT NULL DEFAULT 0,
    attachable INTEGER NOT NULL DEFAULT 0,
    enable_ipv6 INTEGER NOT NULL DEFAULT 0,
    created TEXT NOT NULL,
    labels TEXT,
    cached_at TIMESTAMP NOT NULL,
    PRIMARY KEY (machine_id, name)
);

CREATE INDEX idx_networks_machine_id ON networks(machine_id);
CREATE INDEX idx_networks_cached_at ON networks(cached_at);