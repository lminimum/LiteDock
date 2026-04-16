package database

import "context"

// DB is the common database interface
type DB interface {
	Query(ctx context.Context, q string, args ...any) (interface{}, error)
	QueryRow(ctx context.Context, q string, args ...any) interface{}
	Exec(ctx context.Context, q string, args ...any) error
	Ping(ctx context.Context) error
	Close()
}

// Row is a single row result that can be scanned (for type assertions)
type Row interface {
	Scan(dest ...any) error
}
