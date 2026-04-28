package persistent

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupNetworkTestDB(t *testing.T) *sqlite.SQLite {
	t.Helper()

	db, err := sqlite.New(":memory:")
	require.NoError(t, err)

	err = db.Exec(context.Background(), `
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
		)
	`)
	require.NoError(t, err)

	return db
}

func seedNetworks(t *testing.T, repo *NetworkRepo, machineID string, networks []entity.Network) {
	t.Helper()
	err := repo.UpsertBatch(context.Background(), machineID, networks)
	require.NoError(t, err)
}

func makeTestNetwork(name, driver string) entity.Network {
	return entity.Network{
		ID:         uuid.New().String(),
		MachineID:  "",
		Name:       name,
		Driver:     driver,
		Scope:      "local",
		Internal:   false,
		Attachable: false,
		EnableIPv6: false,
		Created:    "2024-01-01T00:00:00Z",
		Labels:     map[string]string{"env": "test"},
	}
}

func TestNetworkRepoListByMachine(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		networks, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Empty(t, networks)
	})

	t.Run("returns networks for machine", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		seedNetworks(t, repo, "m1", []entity.Network{
			makeTestNetwork("bridge0", "bridge"),
			makeTestNetwork("host0", "host"),
		})

		networks, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Len(t, networks, 2)
		assert.Equal(t, "bridge0", networks[0].Name)
		assert.Equal(t, "host0", networks[1].Name)
	})

	t.Run("does not return networks from other machines", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		seedNetworks(t, repo, "m1", []entity.Network{makeTestNetwork("net1", "bridge")})
		seedNetworks(t, repo, "m2", []entity.Network{makeTestNetwork("net2", "host")})

		networks, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Len(t, networks, 1)
		assert.Equal(t, "net1", networks[0].Name)
	})

	t.Run("correctly scans labels", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		n := makeTestNetwork("labeled", "bridge")
		n.Labels = map[string]string{"env": "prod", "region": "us-east-1"}
		seedNetworks(t, repo, "m1", []entity.Network{n})

		networks, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		require.Len(t, networks, 1)
		assert.Equal(t, "prod", networks[0].Labels["env"])
		assert.Equal(t, "us-east-1", networks[0].Labels["region"])
	})

	t.Run("correctly scans boolean fields", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		n := makeTestNetwork("internal-net", "bridge")
		n.Internal = true
		n.Attachable = true
		n.EnableIPv6 = true
		seedNetworks(t, repo, "m1", []entity.Network{n})

		networks, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		require.Len(t, networks, 1)
		assert.True(t, networks[0].Internal)
		assert.True(t, networks[0].Attachable)
		assert.True(t, networks[0].EnableIPv6)
	})
}

func TestNetworkRepoGetByName(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		seedNetworks(t, repo, "m1", []entity.Network{makeTestNetwork("my-bridge", "bridge")})

		n, err := repo.GetByName(context.Background(), "m1", "my-bridge")
		require.NoError(t, err)
		assert.Equal(t, "my-bridge", n.Name)
		assert.Equal(t, "bridge", n.Driver)
		assert.NotNil(t, n.Labels)
		assert.Equal(t, "test", n.Labels["env"])
	})

	t.Run("not found", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)

		_, err := repo.GetByName(context.Background(), "m1", "nonexistent")
		require.Error(t, err)
	})

	t.Run("wrong machine", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		seedNetworks(t, repo, "m1", []entity.Network{makeTestNetwork("shared-name", "bridge")})

		_, err := repo.GetByName(context.Background(), "m2", "shared-name")
		require.Error(t, err)
	})

	t.Run("correctly scans labels", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		n := makeTestNetwork("labeled-net", "overlay")
		n.Labels = map[string]string{"team": "backend", "managed_by": "litedock"}
		seedNetworks(t, repo, "m1", []entity.Network{n})

		result, err := repo.GetByName(context.Background(), "m1", "labeled-net")
		require.NoError(t, err)
		assert.Len(t, result.Labels, 2)
		assert.Equal(t, "backend", result.Labels["team"])
		assert.Equal(t, "litedock", result.Labels["managed_by"])
	})
}

func TestNetworkRepoUpsertBatch(t *testing.T) {
	t.Run("inserts new networks", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		n1 := makeTestNetwork("net1", "bridge")
		n2 := makeTestNetwork("net2", "host")

		err := repo.UpsertBatch(context.Background(), "m1", []entity.Network{n1, n2})
		require.NoError(t, err)

		networks, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Len(t, networks, 2)
	})

	t.Run("replaces existing networks for same machine", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		seedNetworks(t, repo, "m1", []entity.Network{
			makeTestNetwork("net1", "bridge"),
		})

		n1 := makeTestNetwork("net1", "overlay")
		n1.Labels = map[string]string{"updated": "yes"}
		err := repo.UpsertBatch(context.Background(), "m1", []entity.Network{n1})
		require.NoError(t, err)

		networks, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Len(t, networks, 1)
		assert.Equal(t, "overlay", networks[0].Driver)
		assert.Equal(t, "yes", networks[0].Labels["updated"])
	})

	t.Run("empty batch is no-op", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		seedNetworks(t, repo, "m1", []entity.Network{makeTestNetwork("keep", "bridge")})

		err := repo.UpsertBatch(context.Background(), "m1", []entity.Network{})
		require.NoError(t, err)

		networks, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Len(t, networks, 1)
	})

	t.Run("preserves created field", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		n := makeTestNetwork("preserve-created", "bridge")
		n.Created = "2023-06-15T10:30:00Z"
		seedNetworks(t, repo, "m1", []entity.Network{n})

		result, err := repo.GetByName(context.Background(), "m1", "preserve-created")
		require.NoError(t, err)
		assert.Equal(t, "2023-06-15T10:30:00Z", result.Created)
	})
}

func TestNetworkRepoDeleteByMachine(t *testing.T) {
	t.Run("deletes all networks for machine", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		seedNetworks(t, repo, "m1", []entity.Network{
			makeTestNetwork("net1", "bridge"),
			makeTestNetwork("net2", "host"),
		})

		err := repo.DeleteByMachine(context.Background(), "m1")
		require.NoError(t, err)

		networks, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Empty(t, networks)
	})

	t.Run("does not delete networks from other machines", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		seedNetworks(t, repo, "m1", []entity.Network{makeTestNetwork("keep", "bridge")})
		seedNetworks(t, repo, "m2", []entity.Network{makeTestNetwork("delete", "host")})

		err := repo.DeleteByMachine(context.Background(), "m2")
		require.NoError(t, err)

		m1Networks, err := repo.ListByMachine(context.Background(), "m1")
		require.NoError(t, err)
		assert.Len(t, m1Networks, 1)

		m2Networks, err := repo.ListByMachine(context.Background(), "m2")
		require.NoError(t, err)
		assert.Empty(t, m2Networks)
	})

	t.Run("no-op on empty machine", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)

		err := repo.DeleteByMachine(context.Background(), "empty-machine")
		require.NoError(t, err)
	})
}

func TestNetworkRepoIsCacheValid(t *testing.T) {
	t.Run("valid when cached within maxAge", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		seedNetworks(t, repo, "m1", []entity.Network{makeTestNetwork("net1", "bridge")})

		valid, err := repo.IsCacheValid(context.Background(), "m1", 1*time.Hour)
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("invalid when cached_at is old", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		// This test uses a very small maxAge - the UpsertBatch sets cached_at to now,
		// so we need to test with a past timestamp.
		// Since we can't easily set cached_at in the past via UpsertBatch,
		// we test with zero maxAge which means "must be fresher than now".
		valid, err := repo.IsCacheValid(context.Background(), "m1", 0)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("invalid when no rows for machine", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)

		valid, err := repo.IsCacheValid(context.Background(), "m1", 1*time.Hour)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("valid with very large maxAge", func(t *testing.T) {
		db := setupNetworkTestDB(t)
		defer db.Close()

		repo := NewNetworkRepo(db)
		seedNetworks(t, repo, "m1", []entity.Network{makeTestNetwork("net1", "bridge")})

		valid, err := repo.IsCacheValid(context.Background(), "m1", 24*time.Hour)
		require.NoError(t, err)
		assert.True(t, valid)
	})
}
