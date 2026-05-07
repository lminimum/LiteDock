package persistent

import (
	"context"
	"database/sql"
	"encoding/json"
	stdErrors "errors"
	"time"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/database"
	"github.com/lminimum/LiteDock/pkg/errors"
)

type ImageRepo struct {
	db database.DB
}

func NewImageRepo(db database.DB) *ImageRepo {
	db.Exec(context.Background(), `
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
		)
	`)
	return &ImageRepo{db: db}
}

func (r *ImageRepo) ListByMachine(ctx context.Context, machineID string) ([]entity.Image, error) {
	query := `
		SELECT id, machine_id, repo_tags, repo_digests, size, created_at, labels, cached_at
		FROM images WHERE machine_id = ?`

	rowsInterface, err := r.db.Query(ctx, query, machineID)
	if err != nil {
		return nil, errors.Wrap(err, "ImageRepo.ListByMachine.Query")
	}

	scanner, ok := rowsInterface.(interface {
		Next() bool
		Scan(...any) error
		Close() error
		Err() error
	})
	if !ok {
		return nil, stdErrors.New("ImageRepo.ListByMachine: rows does not implement scanner interface")
	}
	defer scanner.Close()

	images := make([]entity.Image, 0)

	for scanner.Next() {
		var img entity.Image
		var repoTagsJSON, repoDigestsJSON, labelsJSON []byte

		err := scanRow(
			scanner,
			&img.ID,
			&img.MachineID,
			&repoTagsJSON,
			&repoDigestsJSON,
			&img.Size,
			&img.CreatedAt,
			&labelsJSON,
			&img.CachedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "ImageRepo.ListByMachine.scanRow")
		}

		if repoTagsJSON != nil {
			if err := json.Unmarshal(repoTagsJSON, &img.RepoTags); err != nil {
				return nil, errors.Wrap(err, "ImageRepo.ListByMachine.UnmarshalRepoTags")
			}
		}

		if repoDigestsJSON != nil {
			if err := json.Unmarshal(repoDigestsJSON, &img.RepoDigests); err != nil {
				return nil, errors.Wrap(err, "ImageRepo.ListByMachine.UnmarshalRepoDigests")
			}
		}

		if labelsJSON != nil {
			if err := json.Unmarshal(labelsJSON, &img.Labels); err != nil {
				return nil, errors.Wrap(err, "ImageRepo.ListByMachine.UnmarshalLabels")
			}
		}

		images = append(images, img)
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "ImageRepo.ListByMachine.rowsErr")
	}

	return images, nil
}

func (r *ImageRepo) GetByID(ctx context.Context, machineID, imageID string) (*entity.Image, error) {
	query := `
		SELECT id, machine_id, repo_tags, repo_digests, size, created_at, labels, cached_at
		FROM images WHERE machine_id = ? AND id = ?`

	row := r.db.QueryRow(ctx, query, machineID, imageID)

	var img entity.Image
	var repoTagsJSON, repoDigestsJSON, labelsJSON []byte

	err := scanRow(
		row,
		&img.ID,
		&img.MachineID,
		&repoTagsJSON,
		&repoDigestsJSON,
		&img.Size,
		&img.CreatedAt,
		&labelsJSON,
		&img.CachedAt,
	)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrNotFound
		}
		return nil, errors.Wrap(err, "ImageRepo.GetByID.Scan")
	}

	if repoTagsJSON != nil {
		if err := json.Unmarshal(repoTagsJSON, &img.RepoTags); err != nil {
			return nil, errors.Wrap(err, "ImageRepo.GetByID.UnmarshalRepoTags")
		}
	}

	if repoDigestsJSON != nil {
		if err := json.Unmarshal(repoDigestsJSON, &img.RepoDigests); err != nil {
			return nil, errors.Wrap(err, "ImageRepo.GetByID.UnmarshalRepoDigests")
		}
	}

	if labelsJSON != nil {
		if err := json.Unmarshal(labelsJSON, &img.Labels); err != nil {
			return nil, errors.Wrap(err, "ImageRepo.GetByID.UnmarshalLabels")
		}
	}

	return &img, nil
}

func (r *ImageRepo) UpsertBatch(ctx context.Context, machineID string, images []entity.Image) error {
	if len(images) == 0 {
		return nil
	}

	deleteQuery := `DELETE FROM images WHERE machine_id = ?`
	err := r.db.Exec(ctx, deleteQuery, machineID)
	if err != nil {
		return errors.Wrap(err, "ImageRepo.UpsertBatch.Delete")
	}

	insertQuery := `
		INSERT INTO images (id, machine_id, repo_tags, repo_digests, size, created_at, labels, cached_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()

	for _, img := range images {
		repoTagsJSON, err := json.Marshal(img.RepoTags)
		if err != nil {
			return errors.Wrap(err, "ImageRepo.UpsertBatch.MarshalRepoTags")
		}

		repoDigestsJSON, err := json.Marshal(img.RepoDigests)
		if err != nil {
			return errors.Wrap(err, "ImageRepo.UpsertBatch.MarshalRepoDigests")
		}

		labelsJSON, err := json.Marshal(img.Labels)
		if err != nil {
			return errors.Wrap(err, "ImageRepo.UpsertBatch.MarshalLabels")
		}

		err = r.db.Exec(
			ctx, insertQuery,
			img.ID,
			machineID,
			repoTagsJSON,
			repoDigestsJSON,
			img.Size,
			img.CreatedAt,
			labelsJSON,
			now,
		)
		if err != nil {
			return errors.Wrap(err, "ImageRepo.UpsertBatch.Insert")
		}
	}

	return nil
}

func (r *ImageRepo) DeleteByMachine(ctx context.Context, machineID string) error {
	query := `DELETE FROM images WHERE machine_id = ?`

	err := r.db.Exec(ctx, query, machineID)
	if err != nil {
		return errors.Wrap(err, "ImageRepo.DeleteByMachine.Exec")
	}

	return nil
}

func (r *ImageRepo) DeleteByID(ctx context.Context, machineID, imageID string) error {
	query := `DELETE FROM images WHERE machine_id = ? AND id = ?`

	err := r.db.Exec(ctx, query, machineID, imageID)
	if err != nil {
		return errors.Wrap(err, "ImageRepo.DeleteByID.Exec")
	}

	return nil
}

func (r *ImageRepo) IsCacheValid(ctx context.Context, machineID string, maxAge time.Duration) (bool, error) {
	query := `
		SELECT cached_at FROM images WHERE machine_id = ? LIMIT 1`

	row := r.db.QueryRow(ctx, query, machineID)

	var cachedAt time.Time
	err := scanRow(row, &cachedAt)
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, errors.Wrap(err, "ImageRepo.IsCacheValid.QueryRow")
	}

	return time.Now().Before(cachedAt.Add(maxAge)), nil
}

func (r *ImageRepo) CountAll(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM images`

	row := r.db.QueryRow(ctx, query)

	var count int64
	err := scanRow(row, &count)
	if err != nil {
		return 0, errors.Wrap(err, "ImageRepo.CountAll.QueryRow")
	}

	return count, nil
}
