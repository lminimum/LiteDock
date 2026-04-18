package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/lminimum/LiteDock/pkg/errors"
	_ "modernc.org/sqlite" // SQLite database driver
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
	db              *sql.DB
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
		return nil, errors.Wrap(err, "SQLite.New.Open")
	}

	db.SetConnMaxLifetime(s.connMaxLifetime)

	for s.connAttempts > 0 {
		err = db.PingContext(context.Background())
		if err == nil {
			break
		}

		time.Sleep(s.connTimeout)

		s.connAttempts--
	}

	if err != nil {
		return nil, errors.Wrap(err, "SQLite.New.Connect")
	}

	s.db = db

	return s, nil
}

func (s *SQLite) Close() {
	_ = s.db.Close()
}

func (s *SQLite) Query(ctx context.Context, q string, args ...any) (interface{}, error) {
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rows, nil
}

func (s *SQLite) QueryRow(ctx context.Context, q string, args ...any) interface{} {
	return s.db.QueryRowContext(ctx, q, args...)
}

func (s *SQLite) Exec(ctx context.Context, q string, args ...any) error {
	_, err := s.db.ExecContext(ctx, q, args...)

	return err
}

func (s *SQLite) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
