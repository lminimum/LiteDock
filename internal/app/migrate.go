// Package app configures and runs application.
package app

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lminimum/LiteDock/config"
	apperrors "github.com/lminimum/LiteDock/pkg/errors"
)

const (
	_migrateAttempts = 5
	_migrateWait     = 200 * time.Millisecond
)

func AutoMigrate(cfg *config.Config) error {
	dbType := cfg.DB.Type
	if dbType == "" {
		dbType = "postgres"
	}

	databaseURL, err := buildDatabaseURL(dbType, cfg.DB.URL)
	if err != nil {
		return err
	}

	migrationsPath, err := filepath.Abs("migrations")
	if err != nil {
		return fmt.Errorf("AutoMigrate resolve migrations path: %w", err)
	}

	m, err := migrate.New("file://"+migrationsPath, databaseURL)
	if err != nil {
		return fmt.Errorf("AutoMigrate connect: %w", err)
	}
	defer m.Close()

	for i := 0; i < _migrateAttempts; i++ {
		err = m.Up()
		if err == nil || errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		if i < _migrateAttempts-1 {
			time.Sleep(_migrateWait)
		}
	}

	return fmt.Errorf("AutoMigrate %s: %w", dbType, err)
}

func buildDatabaseURL(dbType, url string) (string, error) {
	switch dbType {
	case "postgres":
		if url == "" {
			return "", apperrors.ErrDBURLRequired
		}
		if !strings.Contains(url, "sslmode=") {
			url += "?sslmode=disable"
		}
		return url, nil
	case "mysql":
		if url == "" {
			return "", apperrors.ErrDBURLRequiredMySQL
		}
		if !strings.HasPrefix(url, "mysql://") {
			url = "mysql://" + url
		}
		return url, nil
	case "sqlite":
		if url == "" {
			url = "./data.db"
		}
		if !strings.Contains(url, "://") {
			url = "sqlite://" + url
		}
		return url, nil
	default:
		return "", apperrors.Wrap(apperrors.ErrDBTypeNotSupported, "AutoMigrate."+dbType)
	}
}
