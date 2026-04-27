package persistent

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/sqlite"
)

func setupSystemMetricsTestDB(t *testing.T) *sqlite.SQLite {
	db, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS system_metrics (
			id VARCHAR(36) PRIMARY KEY,
			recorded_at TIMESTAMP,
			cpu_percent DOUBLE PRECISION,
			memory_percent DOUBLE PRECISION,
			disk_percent DOUBLE PRECISION,
			created_at TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

func TestSystemMetricsRepoCreate(t *testing.T) {
	db := setupSystemMetricsTestDB(t)
	defer db.Close()

	repo := NewSystemMetricsRepo(db)
	ctx := context.Background()

	metric := &entity.SystemMetric{
		ID:            uuid.New().String(),
		RecordedAt:    time.Now(),
		CPUPercent:    45.5,
		MemoryPercent: 62.3,
		DiskPercent:   78.1,
	}

	err := repo.Create(ctx, metric)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	history, err := repo.GetHistory(ctx, metric.RecordedAt.Add(-time.Minute))
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(history))
	}

	retrieved := history[0]
	if retrieved.ID != metric.ID {
		t.Errorf("expected id %s, got %s", metric.ID, retrieved.ID)
	}
	if retrieved.CPUPercent != metric.CPUPercent {
		t.Errorf("expected cpu_percent %f, got %f", metric.CPUPercent, retrieved.CPUPercent)
	}
	if retrieved.MemoryPercent != metric.MemoryPercent {
		t.Errorf("expected memory_percent %f, got %f", metric.MemoryPercent, retrieved.MemoryPercent)
	}
	if retrieved.DiskPercent != metric.DiskPercent {
		t.Errorf("expected disk_percent %f, got %f", metric.DiskPercent, retrieved.DiskPercent)
	}
}

func TestSystemMetricsRepoGetHistory(t *testing.T) {
	db := setupSystemMetricsTestDB(t)
	defer db.Close()

	repo := NewSystemMetricsRepo(db)
	ctx := context.Background()

	now := time.Now()

	old := &entity.SystemMetric{
		ID:            uuid.New().String(),
		RecordedAt:    now.Add(-2 * time.Hour),
		CPUPercent:    10.0,
		MemoryPercent: 20.0,
		DiskPercent:   30.0,
	}
	mid := &entity.SystemMetric{
		ID:            uuid.New().String(),
		RecordedAt:    now.Add(-1 * time.Hour),
		CPUPercent:    40.0,
		MemoryPercent: 50.0,
		DiskPercent:   60.0,
	}
	recent := &entity.SystemMetric{
		ID:            uuid.New().String(),
		RecordedAt:    now,
		CPUPercent:    70.0,
		MemoryPercent: 80.0,
		DiskPercent:   90.0,
	}

	for _, m := range []*entity.SystemMetric{old, mid, recent} {
		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	since := now.Add(-90 * time.Minute)
	history, err := repo.GetHistory(ctx, since)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(history) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(history))
	}

	if history[0].ID != mid.ID {
		t.Errorf("expected first metric id %s, got %s", mid.ID, history[0].ID)
	}
	if history[1].ID != recent.ID {
		t.Errorf("expected second metric id %s, got %s", recent.ID, history[1].ID)
	}

	for _, m := range history {
		if m.ID == old.ID {
			t.Errorf("old metric should not be in history")
		}
	}
}

func TestSystemMetricsRepoDeleteOlderThan(t *testing.T) {
	db := setupSystemMetricsTestDB(t)
	defer db.Close()

	repo := NewSystemMetricsRepo(db)
	ctx := context.Background()

	now := time.Now()

	old1 := &entity.SystemMetric{
		ID:            uuid.New().String(),
		RecordedAt:    now.Add(-3 * time.Hour),
		CPUPercent:    10.0,
		MemoryPercent: 20.0,
		DiskPercent:   30.0,
	}
	old2 := &entity.SystemMetric{
		ID:            uuid.New().String(),
		RecordedAt:    now.Add(-2 * time.Hour),
		CPUPercent:    15.0,
		MemoryPercent: 25.0,
		DiskPercent:   35.0,
	}
	recent := &entity.SystemMetric{
		ID:            uuid.New().String(),
		RecordedAt:    now,
		CPUPercent:    70.0,
		MemoryPercent: 80.0,
		DiskPercent:   90.0,
	}

	for _, m := range []*entity.SystemMetric{old1, old2, recent} {
		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	before := now.Add(-1 * time.Hour)
	err := repo.DeleteOlderThan(ctx, before)
	if err != nil {
		t.Fatalf("DeleteOlderThan failed: %v", err)
	}

	history, err := repo.GetHistory(ctx, now.Add(-4*time.Hour))
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("expected 1 metric after delete, got %d", len(history))
	}

	if history[0].ID != recent.ID {
		t.Errorf("expected recent metric id %s, got %s", recent.ID, history[0].ID)
	}
}
