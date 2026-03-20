package sqldbs

// QueryOpts holds optional query clauses for query functions that need
// WHERE conditions and ORDER BY in a single parameter.
type QueryOpts struct {
	WhereIns  []WhereIn
	WhereOps  []WhereOp
	WhereNotNulls []Column
	WhereNulls    []Column
	OrderBys      []OrderBy
}

func (q *QueryOpts) HasWhere() bool {
	return len(q.WhereIns) > 0 || len(q.WhereOps) > 0 || len(q.WhereNotNulls) > 0 || len(q.WhereNulls) > 0
}
