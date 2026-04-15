# Multi-Database Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement a database abstraction layer supporting SQLite, MySQL, and PostgreSQL with a factory pattern for connection management.

**Architecture:** 
- Create a `pkg/database/` package with a common interface `DB` that abstracts all database operations
- Implement factory pattern in `pkg/database/factory.go` to create database connections based on configuration
- Refactor existing `pkg/postgres/` to implement the `DB` interface
- Create new `pkg/mysql/` and `pkg/sqlite/` packages following the same interface
- Update config to select database type via environment variable
- Update repos to depend on the `DB` interface, not concrete implementations

**Tech Stack:** 
- Go 1.25+
- `github.com/jackc/pgx/v5` (PostgreSQL)
- `github.com/go-sql-driver/mysql` (MySQL)
- `modernc.org/sqlite` (SQLite - pure Go)
- `github.com/golang-migrate/migrate/v4` with drivers for each DB

---

## Task 1: Create Database Abstraction Layer

**Files:**
- Create: `pkg/database/database.go` - Common DB interface
- Create: `pkg/database/errors.go` - Common database errors
- Create: `pkg/database/factory.go` - Factory for creating DB connections

**Step 1: Create the database interface**

```go
// pkg/database/database.go
package database

import (
	"context"
	"fmt"
	"time"
)

// DB is the common interface for all database implementations.
type DB interface {
	// Query executes a query that returns rows.
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	// QueryRow executes a query that returns a single row.
	QueryRow(ctx context.Context, sql string, args ...any) Row
	// Exec executes a query that doesn't return rows.
	Exec(ctx context.Context, sql string, args ...any) error
	// Ping checks if the database connection is alive.
	Ping(ctx context.Context) error
	// Close closes the database connection.
	Close() error
	// Builder returns the query builder for this database.
	Builder() StatementBuilder
}

// Row is the interface for a single row result.
type Row interface {
	Scan(dest ...any) error
}

// Rows is the interface for multiple row results.
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
}

// StatementBuilder provides SQL query building capabilities.
type StatementBuilder interface {
	Select(columns ...string) SelectQuery
	Insert(table string) InsertQuery
	Update(table string) UpdateQuery
	Delete(table string) DeleteQuery
}

// SelectQuery builds a SELECT query.
type SelectQuery interface {
	From(table string) SelectQuery
	Where(condition string, args ...any) SelectQuery
	ToSql() (string, []any, error)
}

// InsertQuery builds an INSERT query.
type InsertQuery interface {
	Columns(columns ...string) InsertQuery
	Values(values ...any) InsertQuery
	ToSql() (string, []any, error)
}

// UpdateQuery builds an UPDATE query.
type UpdateQuery interface {
	Set(column string, value any) UpdateQuery
	Where(condition string, args ...any) UpdateQuery
	ToSql() (string, []any, error)
}

// DeleteQuery builds a DELETE query.
type DeleteQuery interface {
	Where(condition string, args ...any) DeleteQuery
	ToSql() (string, []any, error)
}
```

**Step 2: Create common errors**

```go
// pkg/database/errors.go
package database

import "errors"

var (
	ErrNotFound      = errors.New("record not found")
	ErrAlreadyExists = errors.New("record already exists")
	ErrInvalidInput  = errors.New("invalid input")
)
```

**Step 3: Create the factory**

```go
// pkg/database/factory.go
package database

import (
	"fmt"
	"os"
)

// New creates a new database connection based on DB_TYPE environment variable.
// Supported types: postgres, mysql, sqlite
func New() (DB, error) {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "postgres" // default
	}

	switch dbType {
	case "postgres":
		url := os.Getenv("PG_URL")
		if url == "" {
			return nil, fmt.Errorf("PG_URL environment variable is required for postgres")
		}
		return newPostgres(url)
	case "mysql":
		url := os.Getenv("MYSQL_URL")
		if url == "" {
			return nil, fmt.Errorf("MYSQL_URL environment variable is required for mysql")
		}
		return newMySQL(url)
	case "sqlite":
		dsn := os.Getenv("SQLITE_DSN")
		if dsn == "" {
			dsn = "./data.db" // default
		}
		return newSQLite(dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}
```

---

## Task 2: Refactor PostgreSQL Package

**Files:**
- Modify: `pkg/postgres/postgres.go` - Implement DB interface
- Modify: `pkg/postgres/options.go` - Keep existing options

**Step 1: Update postgres.go to implement DB interface**

Replace the entire content of `pkg/postgres/postgres.go` with:

```go
// Package postgres implements postgres database connection.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

// Postgres implements the database.DB interface.
type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	builder squirrel.StatementBuilderType
	pool    *pgxpool.Pool
}

// New creates a new PostgreSQL connection.
func New(url string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	for _, opt := range opts {
		opt(pg)
	}

	pg.builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("postgres - New - pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)

	for pg.connAttempts > 0 {
		pg.pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			break
		}

		time.Sleep(pg.connTimeout)
		pg.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("postgres - New - connAttempts == 0: %w", err)
	}

	return pg, nil
}

// Close closes the database connection.
func (p *Postgres) Close() error {
	if p.pool != nil {
		p.pool.Close()
	}
	return nil
}

// Query executes a query that returns rows.
func (p *Postgres) Query(ctx context.Context, sql string, args ...any) (interface{ Next() bool; Scan(dest ...any) error; Close() error }, error) {
	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// QueryRow executes a query that returns a single row.
func (p *Postgres) QueryRow(ctx context.Context, sql string, args ...any) interface{ Scan(dest ...any) error } {
	return p.pool.QueryRow(ctx, sql, args...)
}

// Exec executes a query that doesn't return rows.
func (p *Postgres) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := p.pool.Exec(ctx, sql, args...)
	return err
}

// Ping checks if the database connection is alive.
func (p *Postgres) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

// Builder returns the query builder.
func (p *Postgres) Builder() interface {
	Select(string) interface {
		From(string) interface {
			Where(string, ...any) interface{ ToSql() (string, []any, error) }
			ToSql() (string, []any, error)
		}
	}
	Insert(string) interface {
		Columns(...string) interface {
			Values(...any) interface{ ToSql() (string, []any, error) }
			ToSql() (string, []any, error)
		}
	}
} {
	return p.builder
}
```

---

## Task 3: Create MySQL Package

**Files:**
- Create: `pkg/mysql/mysql.go` - MySQL implementation
- Create: `pkg/mysql/options.go` - MySQL options

**Step 1: Create mysql.go**

```go
// Package mysql implements MySQL database connection.
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/Masterminds/squirrel"
)

const (
	_defaultMaxOpenConns = 1
	_defaultConnMaxLifetime = time.Minute
)

// MySQL implements the database.DB interface.
type MySQL struct {
	maxOpenConns    int
	connMaxLifetime time.Duration

	builder squirrel.StatementBuilderType
	db      *sql.DB
}

// New creates a new MySQL connection.
func New(dsn string, opts ...Option) (*MySQL, error) {
	m := &MySQL{
		maxOpenConns:    _defaultMaxOpenConns,
		connMaxLifetime: _defaultConnMaxLifetime,
	}

	for _, opt := range opts {
		opt(m)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql - New - sql.Open: %w", err)
	}

	db.SetMaxOpenConns(m.maxOpenConns)
	db.SetConnMaxLifetime(m.connMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("mysql - New - db.Ping: %w", err)
	}

	m.db = db
	m.builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	return m, nil
}

// Close closes the database connection.
func (m *MySQL) Close() error {
	return m.db.Close()
}

// Query executes a query that returns rows.
func (m *MySQL) Query(ctx context.Context, sql string, args ...any) (interface{ Next() bool; Scan(dest ...any) error; Close() error }, error) {
	rows, err := m.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// QueryRow executes a query that returns a single row.
func (m *MySQL) QueryRow(ctx context.Context, sql string, args ...any) interface{ Scan(dest ...any) error } {
	return m.db.QueryRowContext(ctx, sql, args...)
}

// Exec executes a query that doesn't return rows.
func (m *MySQL) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := m.db.ExecContext(ctx, sql, args...)
	return err
}

// Ping checks if the database connection is alive.
func (m *MySQL) Ping(ctx context.Context) error {
	return m.db.PingContext(ctx)
}

// Builder returns the query builder.
func (m *MySQL) Builder() interface {
	Select(string) interface {
		From(string) interface {
			Where(string, ...any) interface{ ToSql() (string, []any, error) }
			ToSql() (string, []any, error)
		}
	}
	Insert(string) interface {
		Columns(...string) interface {
			Values(...any) interface{ ToSql() (string, []any, error) }
			ToSql() (string, []any, error)
		}
	}
} {
	return m.builder
}
```

**Step 2: Create options.go**

```go
package mysql

import "time"

// Option is a functional option for MySQL connection.
type Option func(*MySQL)

// MaxOpenConns sets the maximum number of open connections.
func MaxOpenConns(n int) Option {
	return func(m *MySQL) {
		m.maxOpenConns = n
	}
}

// ConnMaxLifetime sets the maximum lifetime of a connection.
func ConnMaxLifetime(d time.Duration) Option {
	return func(m *MySQL) {
		m.connMaxLifetime = d
	}
}
```

---

## Task 4: Create SQLite Package

**Files:**
- Create: `pkg/sqlite/sqlite.go` - SQLite implementation
- Create: `pkg/sqlite/options.go` - SQLite options

**Step 1: Create sqlite.go**

```go
// Package sqlite implements SQLite database connection.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
	"github.com/Masterminds/squirrel"
)

const (
	_defaultConnMaxLifetime = time.Minute
)

// SQLite implements the database.DB interface.
type SQLite struct {
	connMaxLifetime time.Duration

	builder squirrel.StatementBuilderType
	db      *sql.DB
}

// New creates a new SQLite connection.
func New(dsn string, opts ...Option) (*SQLite, error) {
	s := &SQLite{
		connMaxLifetime: _defaultConnMaxLifetime,
	}

	for _, opt := range opt {
		opt(s)
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite - New - sql.Open: %w", err)
	}

	db.SetConnMaxLifetime(s.connMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("sqlite - New - db.Ping: %w", err)
	}

	s.db = db
	s.builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	return s, nil
}

// Close closes the database connection.
func (s *SQLite) Close() error {
	return s.db.Close()
}

// Query executes a query that returns rows.
func (s *SQLite) Query(ctx context.Context, sql string, args ...any) (interface{ Next() bool; Scan(dest ...any) error; Close() error }, error) {
	rows, err := s.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// QueryRow executes a query that returns a single row.
func (s *SQLite) QueryRow(ctx context.Context, sql string, args ...any) interface{ Scan(dest ...any) error } {
	return s.db.QueryRowContext(ctx, sql, args...)
}

// Exec executes a query that doesn't return rows.
func (s *SQLite) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := s.db.ExecContext(ctx, sql, args...)
	return err
}

// Ping checks if the database connection is alive.
func (s *SQLite) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Builder returns the query builder.
func (s *SQLite) Builder() interface {
	Select(string) interface {
		From(string) interface {
			Where(string, ...any) interface{ ToSql() (string, []any, error) }
			ToSql() (string, []any, error)
		}
	}
	Insert(string) interface {
		Columns(...string) interface {
			Values(...any) interface{ ToSql() (string, []any, error) }
			ToSql() (string, []any, error)
		}
	}
} {
	return s.builder
}
```

**Step 2: Create options.go**

```go
package sqlite

import "time"

// Option is a functional option for SQLite connection.
type Option func(*SQLite)

// ConnMaxLifetime sets the maximum lifetime of a connection.
func ConnMaxLifetime(d time.Duration) Option {
	return func(s *SQLite) {
		s.connMaxLifetime = d
	}
}
```

---

## Task 5: Update Config for Multi-Database

**Files:**
- Modify: `config/config.go` - Add DB config struct

**Step 1: Update config.go**

```go
// PG
type DB struct {
	Type     string `env:"DB_TYPE" envDefault:"postgres"` // postgres, mysql, sqlite
	PoolMax  int    `env:"DB_POOL_MAX,required"`
	URL      string `env:"DB_URL,required"` // Type-specific connection URL
}
```

Remove the old `PG` struct and add `DB` struct. Keep other configs.

---

## Task 6: Update Repos to Use Abstract Interface

**Files:**
- Modify: `internal/repo/persistent/user_postgres.go` → rename to `user_repo.go`
- Modify: `internal/repo/persistent/translation_postgres.go` → rename to `translation_repo.go`
- Modify: `internal/repo/persistent/new.go` → create adapter

**Step 1: Create repo adapter**

Create `internal/repo/persistent/adapter.go`:

```go
package persistent

import (
	"database/sql"
	"fmt"

	"github.com/evrone/go-clean-template/pkg/database"
)

// Adapter wraps a database.DB and provides postgres-compatible methods.
type Adapter struct {
	db database.DB
}

// NewAdapter creates a new database adapter.
func NewAdapter(db database.DB) *Adapter {
	return &Adapter{db: db}
}

// Pool returns the underlying pool (only valid for postgres).
// For other databases, this returns an error.
func (a *Adapter) Pool() (*sql.DB, error) {
	return nil, fmt.Errorf("Pool() is only available for postgres")
}

// DB returns the generic database interface.
func (a *Adapter) DB() database.DB {
	return a.db
}
```

**Step 2: Update user_repo.go to use generic interface**

The UserRepo should use the Adapter pattern and work with the DB interface.

---

## Task 7: Update Migration for Multi-Database

**Files:**
- Modify: `internal/app/migrate.go` - Support multiple database types

**Step 1: Update migrate.go**

```go
//go:build migrate

package app

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	// Database drivers
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
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
	switch dbType {
	case "postgres":
		databaseURL = os.Getenv("PG_URL")
		if databaseURL == "" {
			log.Fatal("migrate: PG_URL environment variable not declared")
		}
		databaseURL += "?sslmode=disable"
	case "mysql":
		databaseURL = os.Getenv("MYSQL_URL")
		if databaseURL == "" {
			log.Fatal("migrate: MYSQL_URL environment variable not declared")
		}
	case "sqlite":
		databaseURL = os.Getenv("SQLITE_DSN")
		if databaseURL == "" {
			databaseURL = "file://./data.db"
		}
	default:
		log.Fatalf("migrate: unsupported database type: %s", dbType)
	}

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
```

---

## Task 8: Update Docker Compose with Multi-Database Services

**Files:**
- Modify: `docker-compose.yml` - Add MySQL and SQLite services

**Step 1: Update docker-compose.yml**

Add MySQL service:

```yaml
  mysql:
    container_name: mysql
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: db
      MYSQL_USER: user
      MYSQL_PASSWORD: myAwEsOm3pa55@w0rd
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    networks:
      app_network:
        aliases:
          - mysql.lvh.me
```

Add SQLite configuration (note: SQLite is file-based, no service needed but configure the path).

Update backend environment:

```yaml
x-backend-app-environment: &x-backend-app-environment
  # ... existing vars ...
  # Database
  DB_TYPE: "postgres"  # postgres, mysql, sqlite
  DB_POOL_MAX: "2"
  DB_URL: "postgres://user:myAwEsOm3pa55@w0rd@db:5432/db"
```

---

## Task 9: Update App to Use New Database Package

**Files:**
- Modify: `internal/app/app.go` - Use database factory

**Step 1: Update app.go**

```go
// Repository
db, err := database.New()
if err != nil {
	l.Fatal(fmt.Errorf("app - Run - database.New: %w", err))
}
defer db.Close()

adapter := persistent.NewAdapter(db)

// Repository instances
persistentRepo := persistent.NewTranslationRepo(adapter)
userRepo := persistent.NewUserRepo(adapter)
```

---

## Task 10: Add Unit Tests

**Files:**
- Create: `pkg/database/database_test.go`
- Create: `pkg/mysql/mysql_test.go`
- Create: `pkg/sqlite/sqlite_test.go`

---

## Execution Options

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach?**
