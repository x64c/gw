package sqldbs

import "context"

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Exec(ctx context.Context, query string, args ...any) (Result, error)
	Query(ctx context.Context, query string, args ...any) (Rows, error)
}
