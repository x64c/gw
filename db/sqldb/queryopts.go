package sqldb

// QueryOpts holds optional query clauses for query functions that need
// both WHERE conditions and ORDER BY in a single parameter.
type QueryOpts struct {
	OrderBys []OrderBy
	Wheres   []WhereEq
}
