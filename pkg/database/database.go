package database

import (
	"context"
)

type DB interface {
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Exec(ctx context.Context, sql string, args ...any) error
	Ping(ctx context.Context) error
	Close() error
	Builder() StatementBuilder
}

type Row interface {
	Scan(dest ...any) error
}

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
}

type StatementBuilder interface {
	Select(columns ...string) SelectQuery
	Insert(table string) InsertQuery
	Update(table string) UpdateQuery
	Delete(table string) DeleteQuery
}

type SelectQuery interface {
	From(table string) SelectQuery
	Where(condition string, args ...any) SelectQuery
	ToSql() (string, []any, error)
}

type InsertQuery interface {
	Columns(columns ...string) InsertQuery
	Values(values ...any) InsertQuery
	ToSql() (string, []any, error)
}

type UpdateQuery interface {
	Set(column string, value any) UpdateQuery
	Where(condition string, args ...any) UpdateQuery
	ToSql() (string, []any, error)
}

type DeleteQuery interface {
	Where(condition string, args ...any) DeleteQuery
	ToSql() (string, []any, error)
}
