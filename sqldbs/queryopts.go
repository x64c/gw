package sqldbs

// QueryOpts holds optional query clauses for query functions that need
// WHERE conditions and ORDER BY in a single parameter.
type QueryOpts struct {
	WhereCond Cond
	OrderBys  []OrderBy
}
