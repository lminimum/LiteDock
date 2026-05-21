package persistent

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/sqlite"
	"github.com/stretchr/testify/require"
)

func setupTaskTestDB(t *testing.T) *sqlite.SQLite {
	t.Helper()

	db, err := sqlite.New(":memory:")
	require.NoError(t, err)

	err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS tasks (
			id VARCHAR(36) PRIMARY KEY,
			type VARCHAR(50) NOT NULL,
			status VARCHAR(20) NOT NULL,
			machine_id VARCHAR(36) NOT NULL,
			payload TEXT,
			result TEXT,
			error TEXT,
			logs TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	require.NoError(t, err)

	return db
}

func TestTaskRepoAppendLogsHandlesNullLogColumn(t *testing.T) {
	db := setupTaskTestDB(t)
	defer db.Close()

	repo := NewTaskRepo(db)
	ctx := context.Background()
	taskID := uuid.New().String()

	err := db.Exec(ctx, `
		INSERT INTO tasks (id, type, status, machine_id, payload, result, error, logs, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, taskID, "container.create", entity.TaskStatusPending, "machine-1", "payload", "", "", nil)
	require.NoError(t, err)

	err = repo.AppendLogs(ctx, taskID, "first line\n")
	require.NoError(t, err)

	task, err := repo.GetByID(ctx, taskID)
	require.NoError(t, err)
	require.Equal(t, "first line\n", task.Logs)
}
