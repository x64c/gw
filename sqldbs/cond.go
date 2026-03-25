package sqldbs

// Cond represents a SQL condition expression.
// Implemented by predicates (BinPred, InPred, NullPred, NotNullPred)
// and logical operators (Not, And, Or).
// BindRepr returns a SQL fragment with ? placeholders and bind args.
// Dialect-specific placeholder translation (e.g. ? → $1) is handled by the final WHERE clause builder.
type Cond interface {
	BindRepr() (string, []any)
}
