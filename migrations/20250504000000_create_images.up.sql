CREATE TABLE IF NOT EXISTS images (
    id TEXT NOT NULL,
    machine_id TEXT NOT NULL,
    repo_tags TEXT NOT NULL DEFAULT '[]',
    repo_digests TEXT NOT NULL DEFAULT '[]',
    size BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    labels TEXT NOT NULL DEFAULT '{}',
    cached_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, machine_id)
);