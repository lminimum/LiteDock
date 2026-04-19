package postgres

import (
	"context"
	stderrors "errors"
	"log"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	apperrors "github.com/lminimum/LiteDock/pkg/errors"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration
	Pool         *pgxpool.Pool
}

var errMaxPoolSizeOutOfRange = stderrors.New("postgres: maxPoolSize out of range")

func New(url string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	for _, opt := range opts {
		opt(pg)
	}

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, apperrors.Wrap(err, "Postgres.New.ParseConfig")
	}

	if pg.maxPoolSize > math.MaxInt32 || pg.maxPoolSize < 0 {
		return nil, errMaxPoolSizeOutOfRange
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)

	for pg.connAttempts > 0 {
		pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			break
		}

		log.Printf("Postgres is trying to connect, attempts left: %d", pg.connAttempts)

		time.Sleep(pg.connTimeout)

		pg.connAttempts--
	}

	if err != nil {
		return nil, apperrors.Wrap(err, "Postgres.New.Connect")
	}

	return pg, nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

func (p *Postgres) Query(ctx context.Context, q string, args ...any) (interface{}, error) {
	rows, err := p.Pool.Query(ctx, rewritePlaceholders(q), args...)

	return rows, err
}

func (p *Postgres) QueryRow(ctx context.Context, q string, args ...any) interface{} {
	return p.Pool.QueryRow(ctx, rewritePlaceholders(q), args...)
}

func (p *Postgres) Exec(ctx context.Context, q string, args ...any) error {
	_, err := p.Pool.Exec(ctx, rewritePlaceholders(q), args...)

	return err
}

func (p *Postgres) Ping(ctx context.Context) error {
	return p.Pool.Ping(ctx)
}
