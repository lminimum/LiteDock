package persistent

import (
	"context"
	"testing"
	"time"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupVolumeTestDB(t *testing.T) *sqlite.SQLite {
	t.Helper()

	db, err := sqlite.New(":memory:")
	require.NoError(t, err)

	err = db.Exec(context.Background(), `
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
		)
	`)
	require.NoError(t, err)

	return db
}

func seedVolumes(t *testing.T, repo *VolumeRepo, machineID string, volumes []entity.Volume) {
	t.Helper()
	err := repo.UpsertBatch(context.Background(), machineID, volumes)
	require.NoError(t, err)
}

func makeTestVolume(name, driver string) entity.Volume {
	return entity.Volume{
		Name:       name,
		MachineID:  "",
		Driver:     driver,
		Mountpoint: "/var/lib/docker/volumes/" + name + "/_data",
		CreatedAt:  "2024-01-01T00:00:00Z",
		Scope:      "local",
		Labels:     map[string]string{"env": "test"},
		Size:       1024,
	}
}

func TestVolumeRepoListByMachine(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		volumes, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Empty(t, volumes)
	})

	t.Run("returns volumes for machine", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		seedVolumes(t, repo, "m1", []entity.Volume{
			makeTestVolume("vol1", "local"),
			makeTestVolume("vol2", "local"),
		})

		volumes, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Len(t, volumes, 2)
		assert.Equal(t, "vol1", volumes[0].Name)
		assert.Equal(t, "vol2", volumes[1].Name)
	})

	t.Run("does not return volumes from other machines", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		seedVolumes(t, repo, "m1", []entity.Volume{makeTestVolume("vol1", "local")})
		seedVolumes(t, repo, "m2", []entity.Volume{makeTestVolume("vol2", "local")})

		volumes, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Len(t, volumes, 1)
		assert.Equal(t, "vol1", volumes[0].Name)
	})

	t.Run("correctly scans labels", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		v := makeTestVolume("labeled-vol", "local")
		v.Labels = map[string]string{"env": "prod", "region": "us-east-1"}
		seedVolumes(t, repo, "m1", []entity.Volume{v})

		volumes, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		require.Len(t, volumes, 1)
		assert.Equal(t, "prod", volumes[0].Labels["env"])
		assert.Equal(t, "us-east-1", volumes[0].Labels["region"])
	})

	t.Run("correctly scans size field", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		v := makeTestVolume("big-vol", "local")
		v.Size = 4096
		seedVolumes(t, repo, "m1", []entity.Volume{v})

		volumes, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		require.Len(t, volumes, 1)
		assert.Equal(t, int64(4096), volumes[0].Size)
	})
}

func TestVolumeRepoGetByName(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		seedVolumes(t, repo, "m1", []entity.Volume{makeTestVolume("my-volume", "local")})

		v, err := repo.GetByName(context.Background(), "m1", "my-volume")
		require.NoError(t, err)
		assert.Equal(t, "my-volume", v.Name)
		assert.Equal(t, "local", v.Driver)
		assert.NotNil(t, v.Labels)
		assert.Equal(t, "test", v.Labels["env"])
		assert.Equal(t, int64(1024), v.Size)
	})

	t.Run("not found", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)

		_, err := repo.GetByName(context.Background(), "m1", "nonexistent")
		require.Error(t, err)
	})

	t.Run("wrong machine", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		seedVolumes(t, repo, "m1", []entity.Volume{makeTestVolume("shared-name", "local")})

		_, err := repo.GetByName(context.Background(), "m2", "shared-name")
		require.Error(t, err)
	})

	t.Run("correctly scans labels", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		v := makeTestVolume("labeled-vol", "local")
		v.Labels = map[string]string{"team": "backend", "managed_by": "litedock"}
		seedVolumes(t, repo, "m1", []entity.Volume{v})

		result, err := repo.GetByName(context.Background(), "m1", "labeled-vol")
		require.NoError(t, err)
		assert.Len(t, result.Labels, 2)
		assert.Equal(t, "backend", result.Labels["team"])
		assert.Equal(t, "litedock", result.Labels["managed_by"])
	})
}

func TestVolumeRepoUpsertBatch(t *testing.T) {
	t.Run("inserts new volumes", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		v1 := makeTestVolume("vol1", "local")
		v2 := makeTestVolume("vol2", "local")

		err := repo.UpsertBatch(context.Background(), "m1", []entity.Volume{v1, v2})
		require.NoError(t, err)

		volumes, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Len(t, volumes, 2)
	})

	t.Run("replaces existing volumes for same machine", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		seedVolumes(t, repo, "m1", []entity.Volume{
			makeTestVolume("vol1", "local"),
		})

		v1 := makeTestVolume("vol1", "nfs")
		v1.Labels = map[string]string{"updated": "yes"}
		v1.Size = 8192
		err := repo.UpsertBatch(context.Background(), "m1", []entity.Volume{v1})
		require.NoError(t, err)

		volumes, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Len(t, volumes, 1)
		assert.Equal(t, "nfs", volumes[0].Driver)
		assert.Equal(t, "yes", volumes[0].Labels["updated"])
		assert.Equal(t, int64(8192), volumes[0].Size)
	})

	t.Run("empty batch is no-op", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		seedVolumes(t, repo, "m1", []entity.Volume{makeTestVolume("keep", "local")})

		err := repo.UpsertBatch(context.Background(), "m1", []entity.Volume{})
		require.NoError(t, err)

		volumes, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Len(t, volumes, 1)
	})

	t.Run("preserves created_at field", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		v := makeTestVolume("preserve-created", "local")
		v.CreatedAt = "2023-06-15T10:30:00Z"
		seedVolumes(t, repo, "m1", []entity.Volume{v})

		result, err := repo.GetByName(context.Background(), "m1", "preserve-created")
		require.NoError(t, err)
		assert.Equal(t, "2023-06-15T10:30:00Z", result.CreatedAt)
	})
}

func TestVolumeRepoDeleteByMachine(t *testing.T) {
	t.Run("deletes all volumes for machine", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		seedVolumes(t, repo, "m1", []entity.Volume{
			makeTestVolume("vol1", "local"),
			makeTestVolume("vol2", "local"),
		})

		err := repo.DeleteByMachine(context.Background(), "m1")
		require.NoError(t, err)

		volumes, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Empty(t, volumes)
	})

	t.Run("does not delete volumes from other machines", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		seedVolumes(t, repo, "m1", []entity.Volume{makeTestVolume("keep", "local")})
		seedVolumes(t, repo, "m2", []entity.Volume{makeTestVolume("delete", "local")})

		err := repo.DeleteByMachine(context.Background(), "m2")
		require.NoError(t, err)

		m1Volumes, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Len(t, m1Volumes, 1)

		m2Volumes, err := repo.ListByMachine(context.Background(), "m2")
		require.NoError(t, err)
		assert.Empty(t, m2Volumes)
	})

	t.Run("no-op on empty machine", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)

		err := repo.DeleteByMachine(context.Background(), "empty-machine")
		require.NoError(t, err)
	})
}

func TestVolumeRepoIsCacheValid(t *testing.T) {
	t.Run("valid when cached within maxAge", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		seedVolumes(t, repo, "m1", []entity.Volume{makeTestVolume("vol1", "local")})

		valid, err := repo.IsCacheValid(context.Background(), "m1", 1*time.Hour)
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("invalid when cached_at is old", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		seedVolumes(t, repo, "m1", []entity.Volume{makeTestVolume("vol1", "local")})

		valid, err := repo.IsCacheValid(context.Background(), "m1", 0)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("invalid when no rows for machine", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)

		valid, err := repo.IsCacheValid(context.Background(), "m1", 1*time.Hour)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("valid with very large maxAge", func(t *testing.T) {
		db := setupVolumeTestDB(t)
		defer db.Close()

		repo := NewVolumeRepo(db)
		seedVolumes(t, repo, "m1", []entity.Volume{makeTestVolume("vol1", "local")})

		valid, err := repo.IsCacheValid(context.Background(), "m1", 24*time.Hour)
		require.NoError(t, err)
		assert.True(t, valid)
	})
}
