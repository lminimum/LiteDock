package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lminimum/LiteDock/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	_defaultAttempts = 20
	_defaultTimeout  = time.Second
)

func AutoMigrate(cfg *config.Config) error {
	dbType := cfg.DB.Type
	if dbType == "" {
		dbType = "postgres"
	}

	var databaseURL string

	switch dbType {
	case "postgres":
		databaseURL = cfg.DB.URL
		if databaseURL == "" {
			return fmt.Errorf("migrate: DB.URL is required for postgres")
		}
		if !strings.Contains(databaseURL, "sslmode=") {
			databaseURL += "?sslmode=disable"
		}
	case "mysql":
		databaseURL = cfg.DB.URL
		if databaseURL == "" {
			return fmt.Errorf("migrate: DB.URL is required for mysql")
		}
		if !strings.HasPrefix(databaseURL, "mysql://") {
			databaseURL = "mysql://" + databaseURL
		}
	case "sqlite":
		databaseURL = cfg.DB.URL
		if databaseURL == "" {
			databaseURL = "./data.db"
		}
		if !strings.Contains(databaseURL, "://") {
			databaseURL = "sqlite://" + databaseURL
		}
	default:
		return fmt.Errorf("migrate: unsupported database type: %s", dbType)
	}

	var (
		attempts = _defaultAttempts
		err     error
		m       *migrate.Migrate
	)

	for attempts > 0 {
		m, err = migrate.New("file://migrations", databaseURL)
		if err == nil {
			break
		}

		time.Sleep(_defaultTimeout)
		attempts--
	}

	if err != nil {
		return fmt.Errorf("migrate: %s connect error: %w", dbType, err)
	}

	err = m.Up()
	defer m.Close()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate: up error: %w", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}

	return nil
}
