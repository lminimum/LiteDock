package database

import (
	"fmt"
	"os"

	"github.com/evrone/go-clean-template/pkg/mysql"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/evrone/go-clean-template/pkg/sqlite"
)

func NewPostgres() (*postgres.Postgres, error) {
	url := os.Getenv("PG_URL")
	if url == "" {
		return nil, fmt.Errorf("PG_URL environment variable is required")
	}
	return postgres.New(url)
}

func NewMySQL() (*mysql.MySQL, error) {
	url := os.Getenv("MYSQL_URL")
	if url == "" {
		return nil, fmt.Errorf("MYSQL_URL environment variable is required")
	}
	return mysql.New(url)
}

func NewSQLite() (*sqlite.SQLite, error) {
	dsn := os.Getenv("SQLITE_DSN")
	if dsn == "" {
		dsn = "./data.db"
	}
	return sqlite.New(dsn)
}

func New() (DB, error) {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "postgres"
	}

	switch dbType {
	case "postgres":
		return NewPostgres()
	case "mysql":
		return NewMySQL()
	case "sqlite":
		return NewSQLite()
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}
