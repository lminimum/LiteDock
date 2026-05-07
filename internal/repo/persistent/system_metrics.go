package persistent

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/database"
)

type SystemMetricsRepo struct {
	db database.DB
}

func NewSystemMetricsRepo(db database.DB) *SystemMetricsRepo {
	return &SystemMetricsRepo{db: db}
}

func (r *SystemMetricsRepo) Create(ctx context.Context, m *entity.SystemMetric) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}

	query := `
		INSERT INTO system_metrics (id, recorded_at, cpu_percent, memory_percent, disk_percent, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	now := time.Now()
	err := r.db.Exec(ctx, query, m.ID, m.RecordedAt, m.CPUPercent, m.MemoryPercent, m.DiskPercent, now)
	if err != nil {
		return err
	}

	return nil
}

func (r *SystemMetricsRepo) GetHistory(ctx context.Context, since time.Time) ([]entity.SystemMetric, error) {
	query := `
		SELECT id, recorded_at, cpu_percent, memory_percent, disk_percent, created_at
		FROM system_metrics
		WHERE recorded_at >= ?
		ORDER BY recorded_at ASC`

	rowsInterface, err := r.db.Query(ctx, query, since)
	if err != nil {
		return nil, err
	}

	scanner, ok := rowsInterface.(interface {
		Next() bool
		Scan(...any) error
		Close() error
		Err() error
	})
	if !ok {
		return nil, errors.New("SystemMetricsRepo.GetHistory: rows does not implement scanner interface")
	}
	defer scanner.Close()

	metrics := make([]entity.SystemMetric, 0)

	for scanner.Next() {
		var m entity.SystemMetric
		err := scanRow(
			scanner,
			&m.ID,
			&m.RecordedAt,
			&m.CPUPercent,
			&m.MemoryPercent,
			&m.DiskPercent,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (r *SystemMetricsRepo) DeleteOlderThan(ctx context.Context, before time.Time) error {
	query := `DELETE FROM system_metrics WHERE recorded_at < ?`

	err := r.db.Exec(ctx, query, before)
	if err != nil {
		return err
	}

	return nil
}

func (r *SystemMetricsRepo) PruneToCount(ctx context.Context, maxCount int) error {
	query := `
		DELETE FROM system_metrics
		WHERE id NOT IN (
			SELECT id FROM system_metrics
			ORDER BY recorded_at DESC
			LIMIT ?
		)`

	err := r.db.Exec(ctx, query, maxCount)
	if err != nil {
		return err
	}

	return nil
}
