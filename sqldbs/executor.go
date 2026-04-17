package sqldbs

import "context"

// Executor is the shared interface between DB and Tx for executing SQL.
// Use this as the parameter type when the caller should work with either.
type Executor interface {
	Exec(ctx context.Context, query string, args ...any) (Result, error)
	Client() Client

	// Query (SELECT columns FROM table ...)

	QueryRow(ctx context.Context, table string, columns []string, id any) Row
	QueryRows(ctx context.Context, table string, columns []string, where Cond) (Rows, error)
	QueryRowRaw(ctx context.Context, query string, args ...any) Row
	QueryRowsRaw(ctx context.Context, query string, args ...any) (Rows, error)

	// Insert (INSERT INTO table (columns) VALUES (values) ...)

	InsertRow(ctx context.Context, table string, columns []string, values []any) (Result, error)
	InsertRows(ctx context.Context, table string, columns []string, rowValues [][]any) (int64, error)
	InsertRowsRaw(ctx context.Context, query string, args ...any) (Result, error)

	// Update (UPDATE table SET column = value, ... WHERE ...)

	UpdateRow(ctx context.Context, table string, pkColumn string, id any, columns []string, values []any) (Result, error)
	UpdateRows(ctx context.Context, table string, columns []string, values []any, where Cond) (int64, error)
	UpdateRowsRaw(ctx context.Context, query string, args ...any) (Result, error)

	// Delete (DELETE FROM table WHERE ...)

	DeleteRow(ctx context.Context, table string, pkColumn string, id any) (Result, error)
	DeleteRows(ctx context.Context, table string, where Cond) (int64, error)
	DeleteRowsRaw(ctx context.Context, query string, args ...any) (Result, error)
}
