package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	_ "modernc.org/sqlite"
)

const (
	_defaultConnMaxLifetime = time.Minute
	_defaultConnAttempts    = 10
	_defaultConnTimeout     = time.Second
)

type SQLite struct {
	connMaxLifetime time.Duration
	connAttempts    int
	connTimeout     time.Duration

	builder squirrel.StatementBuilderType
	db      *sql.DB
}

func New(dsn string, opts ...Option) (*SQLite, error) {
	s := &SQLite{
		connMaxLifetime: _defaultConnMaxLifetime,
		connAttempts:    _defaultConnAttempts,
		connTimeout:     _defaultConnTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite - New - sql.Open: %w", err)
	}

	db.SetConnMaxLifetime(s.connMaxLifetime)

	for s.connAttempts > 0 {
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(s.connTimeout)
		s.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("sqlite - New - connAttempts == 0: %w", err)
	}

	s.db = db
	s.builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	return s, nil
}

func (s *SQLite) Close() {
	_ = s.db.Close()
}

func (s *SQLite) Query(ctx context.Context, sql string, args ...any) (interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
}, error) {
	rows, err := s.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *SQLite) QueryRow(ctx context.Context, sql string, args ...any) interface{ Scan(dest ...any) error } {
	return s.db.QueryRowContext(ctx, sql, args...)
}

func (s *SQLite) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := s.db.ExecContext(ctx, sql, args...)
	return err
}

func (s *SQLite) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *SQLite) Builder() squirrel.StatementBuilderType {
	return s.builder
}
