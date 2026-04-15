//go:build migrate

package app

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

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

func init() {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "postgres"
	}

	var databaseURL string
	var driverName string

	switch dbType {
	case "postgres":
		databaseURL = os.Getenv("PG_URL")
		if databaseURL == "" {
			log.Fatal("migrate: environment variable not declared: PG_URL")
		}
		databaseURL += "?sslmode=disable"
		driverName = "postgres"
	case "mysql":
		databaseURL = os.Getenv("MYSQL_URL")
		if databaseURL == "" {
			log.Fatal("migrate: environment variable not declared: MYSQL_URL")
		}
		driverName = "mysql"
	case "sqlite":
		databaseURL = os.Getenv("SQLITE_DSN")
		if databaseURL == "" {
			databaseURL = "./data.db"
		}
		driverName = "sqlite"
	default:
		log.Fatalf("migrate: unsupported database type: %s", dbType)
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

		log.Printf("Migrate: %s is trying to connect, attempts left: %d", dbType, attempts)
		time.Sleep(_defaultTimeout)
		attempts--
	}

	if err != nil {
		log.Fatalf("Migrate: %s connect error: %s", dbType, err)
	}

	err = m.Up()
	defer m.Close()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Migrate: up error: %s", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Printf("Migrate: no change")
		return
	}

	log.Printf("Migrate: up success")
}
