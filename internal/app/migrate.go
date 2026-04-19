package app

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"    // MySQL driver for migrations
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Postgres driver for migrations
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"   // SQLite driver for migrations
	_ "github.com/golang-migrate/migrate/v4/source/file"       // File source driver for migrations
	"github.com/lminimum/LiteDock/config"
	apperrors "github.com/lminimum/LiteDock/pkg/errors"
)

const (
	_defaultAttempts = 20
	_defaultTimeout  = time.Second
)

const _defaultDBType = "postgres"

func AutoMigrate(cfg *config.Config) error {
	dbType := cfg.DB.Type
	if dbType == "" {
		dbType = _defaultDBType
	}

	databaseURL, err := buildDatabaseURL(dbType, cfg.DB.URL)
	if err != nil {
		return err
	}

	m, err := connectMigrate(databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	return runMigrations(m, dbType)
}

func buildDatabaseURL(dbType, url string) (string, error) {
	switch dbType {
	case "postgres":
		return buildPostgresURL(url)
	case "mysql":
		return buildMySQLURL(url)
	case "sqlite":
		return buildSQLiteURL(url)
	default:
		return "", apperrors.Wrap(apperrors.ErrDBTypeNotSupported, "AutoMigrate."+dbType)
	}
}

func buildPostgresURL(url string) (string, error) {
	if url == "" {
		return "", apperrors.ErrDBURLRequired
	}

	if !strings.Contains(url, "sslmode=") {
		url += "?sslmode=disable"
	}

	return url, nil
}

func buildMySQLURL(url string) (string, error) {
	if url == "" {
		return "", apperrors.ErrDBURLRequiredMySQL
	}

	if !strings.HasPrefix(url, "mysql://") {
		url = "mysql://" + url
	}

	return url, nil
}

func buildSQLiteURL(url string) (string, error) {
	if url == "" {
		url = "./data.db"
	}

	if !strings.Contains(url, "://") {
		url = "sqlite://" + url
	}

	return url, nil
}

func connectMigrate(databaseURL string) (*migrate.Migrate, error) {
	var (
		attempts = _defaultAttempts
		err      error
		m        *migrate.Migrate
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
		return nil, apperrors.Wrap(err, "connectMigrate")
	}

	return m, nil
}

func runMigrations(m *migrate.Migrate, dbType string) error {
	err := m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return apperrors.Wrap(err, "AutoMigrate."+dbType+".Up")
	}

	return nil
}
