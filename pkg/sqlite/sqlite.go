package sqlite

import (
	"context"
	"database/sql"
	"strings"
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

	// Ensure WAL journal mode and busy timeout for concurrent access.
	// Without WAL, SQLite locks the entire database on writes, blocking all reads.
	// Without busy_timeout, concurrent writes immediately fail with SQLITE_BUSY.
	dsn = appendDSNParams(dsn)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "SQLite.New.Open")
	}

	db.SetConnMaxLifetime(s.connMaxLifetime)
	db.SetMaxOpenConns(1)

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

	if err := rows.Err(); err != nil {
		rows.Close()
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

// appendDSNParams adds _journal_mode=WAL and _busy_timeout=5000 for concurrent access.
// WAL allows concurrent reads during writes. busy_timeout retries instead of SQLITE_BUSY.
func appendDSNParams(dsn string) string {
	params := "_journal_mode=WAL&_busy_timeout=5000"

	if dsn == "" {
		return "file:data.db?" + params
	}

	if dsn[len(dsn)-1] == ')' || dsn[len(dsn)-1] == '>' {
		return dsn
	}

	if idx := strings.Index(dsn, "?"); idx >= 0 {
		return dsn + "&" + params
	}

	return dsn + "?" + params
}
