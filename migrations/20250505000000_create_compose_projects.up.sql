CREATE TABLE IF NOT EXISTS compose_projects (
    id UUID PRIMARY KEY,
    machine_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    project_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'unknown',
    services JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    cached_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(machine_id, project_name)
);