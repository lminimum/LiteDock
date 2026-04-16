package app

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/evrone/go-clean-template/config"
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
	var driverName string

	switch dbType {
	case "postgres":
		databaseURL = cfg.DB.URL
		if databaseURL == "" {
			return fmt.Errorf("migrate: DB.URL is required for postgres")
		}
		databaseURL += "?sslmode=disable"
		driverName = "postgres"
	case "mysql":
		databaseURL = os.Getenv("MYSQL_URL")
		if databaseURL == "" {
			return fmt.Errorf("migrate: MYSQL_URL environment variable is required")
		}
		driverName = "mysql"
	case "sqlite":
		databaseURL = cfg.DB.SQLiteDSN
		if databaseURL == "" {
			databaseURL = "./data.db"
		}
		driverName = "sqlite"
	default:
		return fmt.Errorf("migrate: unsupported database type: %s", dbType)
	}

	var (
		attempts = _defaultAttempts
		err      error
		m        *migrate.Migrate
	)

	for attempts > 0 {
		m, err = migrate.New("file://migrations", fmt.Sprintf("%s://%s", driverName, databaseURL))
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
