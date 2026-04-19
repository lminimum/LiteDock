package database

import (
	"os"
	"strings"

	"github.com/lminimum/LiteDock/pkg/errors"
	"github.com/lminimum/LiteDock/pkg/mysql"
	"github.com/lminimum/LiteDock/pkg/postgres"
	"github.com/lminimum/LiteDock/pkg/sqlite"
)

func NewPostgres() (*postgres.Postgres, error) {
	url := os.Getenv("DB_URL")
	if url == "" {
		return nil, errors.ErrDBURLRequired
	}

	return postgres.New(url)
}

func NewMySQL() (*mysql.MySQL, error) {
	url := os.Getenv("DB_URL")
	if url == "" {
		return nil, errors.ErrDBURLRequiredMySQL
	}

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

func convertMySQLURLToDSN(url string) string {
	if !strings.HasPrefix(url, "mysql://") {
		return url
	}

	url = strings.TrimPrefix(url, "mysql://")

	return url
}

const _defaultDBType = "postgres"

func New() (DB, error) {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = _defaultDBType
	}

	switch dbType {
	case "postgres":
		return NewPostgres()
	case "mysql":
		return NewMySQL()
	case "sqlite":
		return NewSQLite()
	default:
		return nil, errors.Wrap(errors.ErrDBTypeNotSupported, "Database.New."+dbType)
	}
}
