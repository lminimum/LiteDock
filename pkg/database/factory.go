package database

import (
	"fmt"
	"os"
	"strings"

	"github.com/lminimum/LiteDock/pkg/mysql"
	"github.com/lminimum/LiteDock/pkg/postgres"
	"github.com/lminimum/LiteDock/pkg/sqlite"
)

func NewPostgres() (*postgres.Postgres, error) {
	url := os.Getenv("DB_URL")
	if url == "" {
		return nil, fmt.Errorf("DB_URL environment variable is required for postgres")
	}
	return postgres.New(url)
}

func NewMySQL() (*mysql.MySQL, error) {
	url := os.Getenv("DB_URL")
	if url == "" {
		return nil, fmt.Errorf("DB_URL environment variable is required for mysql")
	}
	// go-sql-driver uses DSN format: user:pass@tcp(host:port)/db
	// but migrate expects URL format: mysql://user:pass@tcp(host:port)/db
	// Convert URL format to DSN format
	dsn := convertMySQLURLToDSN(url)
	return mysql.New(dsn)
}

func NewSQLite() (*sqlite.SQLite, error) {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		dsn = "./data.db"
	}
	return sqlite.New(dsn)
}

// convertMySQLURLToDSN converts mysql:// URLs to go-sql-driver DSN format
func convertMySQLURLToDSN(url string) string {
	if !strings.HasPrefix(url, "mysql://") {
		return url // already in DSN format
	}
	// url format: mysql://user:pass@tcp(host:port)/db
	// dsn format: user:pass@tcp(host:port)/db
	url = strings.TrimPrefix(url, "mysql://")
	return url
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
