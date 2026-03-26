package sqldbs

import "context"

type DB interface {
	Exec(ctx context.Context, query string, args ...any) (Result, error)
	QueryRows(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) Row
	InsertStmt(ctx context.Context, query string, args ...any) (Result, error)
	CopyFrom(ctx context.Context, table string, columns []string, rows [][]any) (int64, error)
	Listen(ctx context.Context, channel string) (<-chan Notification, error)
	Prepare(ctx context.Context, query string) (PreparedStmt, error)
	BeginTx(ctx context.Context) (Tx, error)
	Ping(ctx context.Context) error

	// Back-reference to the parent Client

	Client() Client

	// Convenience: delegates to Client

	FirstPlaceholder() string
	NthPlaceholder(n int) string
	InPlaceholders(start, cnt int) string
	RawSQLStore(name string) *RawSQLStore

	// Per-DB primary store shortcut (set by app at boot)

	MainRawSQLStore() *RawSQLStore
	SetMainRawSQLStore(name string)
}
