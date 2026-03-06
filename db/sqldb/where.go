package sqldb

import "strings"

// WhereEq defines a validated WHERE column = ? clause.
type WhereEq struct {
	Column Column
	Value  any
}

// WhereEqClause builds " AND col1 = ? AND col2 = ?" from WhereEq entries.
// startNth is the placeholder numbering offset (for PostgreSQL $n style).
// Returns the SQL fragment and the bind values.
func WhereEqClause(wheres []WhereEq, dbClient Client, startNth int) (string, []any) {
	if len(wheres) == 0 {
		return "", nil
	}
	var b strings.Builder
	vals := make([]any, len(wheres))
	for i, w := range wheres {
		b.WriteString(" AND ")
		b.WriteString(w.Column.Name())
		b.WriteString(" = ")
		b.WriteString(dbClient.SinglePlaceholder(startNth + i))
		vals[i] = w.Value
	}
	return b.String(), vals
}
