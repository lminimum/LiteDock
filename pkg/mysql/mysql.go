package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"
)

const (
	_defaultMaxOpenConns    = 1
	_defaultConnMaxLifetime = time.Minute
	_defaultConnAttempts    = 10
	_defaultConnTimeout     = time.Second
)

type MySQL struct {
	maxOpenConns    int
	connMaxLifetime time.Duration
	connAttempts    int
	connTimeout     time.Duration

	builder squirrel.StatementBuilderType
	db      *sql.DB
}

func New(dsn string, opts ...Option) (*MySQL, error) {
	m := &MySQL{
		maxOpenConns:    _defaultMaxOpenConns,
		connMaxLifetime: _defaultConnMaxLifetime,
		connAttempts:    _defaultConnAttempts,
		connTimeout:     _defaultConnTimeout,
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

	for m.connAttempts > 0 {
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(m.connTimeout)
		m.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("mysql - New - connAttempts == 0: %w", err)
	}

	m.db = db
	m.builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	return m, nil
}

func (m *MySQL) Close() {
	_ = m.db.Close()
}

func (m *MySQL) Query(ctx context.Context, sql string, args ...any) (interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
}, error) {
	rows, err := m.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *MySQL) QueryRow(ctx context.Context, sql string, args ...any) interface{ Scan(dest ...any) error } {
	return m.db.QueryRowContext(ctx, sql, args...)
}

func (m *MySQL) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := m.db.ExecContext(ctx, sql, args...)
	return err
}

func (m *MySQL) Ping(ctx context.Context) error {
	return m.db.PingContext(ctx)
}

func (m *MySQL) Builder() squirrel.StatementBuilderType {
	return m.builder
}
