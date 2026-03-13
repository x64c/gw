package sqldbs

import "strings"

type CompareOp struct {
	op string
}

var (
	OpEq   = CompareOp{" = "}
	OpGt   = CompareOp{" > "}
	OpLt   = CompareOp{" < "}
	OpGtEq = CompareOp{" >= "}
	OpLtEq = CompareOp{" <= "}
)

type WhereOpCond struct {
	Column Column
	Op     CompareOp
	Value  any
}

// CompoundWhereOpCond builds " AND col1 = ? AND col2 > ?" from WhereOpCond entries.
// Intended to be appended after an existing WHERE clause (e.g. WHERE fk IN (?)).
// startNth is the placeholder numbering offset (for PostgreSQL $n style).
// Returns the SQL fragment and the bind values.
func CompoundWhereOpCond(wheres []WhereOpCond, db DB, startNth int) (string, []any) {
	if len(wheres) == 0 {
		return "", nil
	}
	var b strings.Builder
	vals := make([]any, len(wheres))
	for i, w := range wheres {
		b.WriteString(" AND ")
		b.WriteString(w.Column.Name())
		b.WriteString(w.Op.op)
		b.WriteString(db.SinglePlaceholder(startNth + i))
		vals[i] = w.Value
	}
	return b.String(), vals
}

// CompoundWhereNotNullCond builds " AND col1 IS NOT NULL AND col2 IS NOT NULL".
// Intended to be appended after an existing WHERE clause (e.g. WHERE fk IN (?)).
// No bind values needed.
func CompoundWhereNotNullCond(columns []Column) string {
	if len(columns) == 0 {
		return ""
	}
	var b strings.Builder
	for _, c := range columns {
		b.WriteString(" AND ")
		b.WriteString(c.Name())
		b.WriteString(" IS NOT NULL")
	}
	return b.String()
}

// CompoundWhereNullCond builds " AND col1 IS NULL AND col2 IS NULL".
// Intended to be appended after an existing WHERE clause (e.g. WHERE fk IN (?)).
// No bind values needed.
func CompoundWhereNullCond(columns []Column) string {
	if len(columns) == 0 {
		return ""
	}
	var b strings.Builder
	for _, c := range columns {
		b.WriteString(" AND ")
		b.WriteString(c.Name())
		b.WriteString(" IS NULL")
	}
	return b.String()
}
