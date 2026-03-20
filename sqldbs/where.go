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

type WhereIn struct {
	Column Column
	Values []any
}

type WhereOp struct {
	Column Column
	Op     CompareOp
	Value  any
}

// CompoundWhereIn builds " AND col1 IN (?, ?) AND col2 IN (?, ?, ?)" from WhereIn entries.
// Intended to be appended after an existing WHERE clause.
// startNth is the placeholder numbering offset (e.g. PostgreSQL $n style).
// Returns the SQL fragment and the bind values.
func CompoundWhereIn(wheres []WhereIn, db DB, startNth int) (string, []any) {
	if len(wheres) == 0 {
		return "", nil
	}
	var b strings.Builder
	var vals []any
	nth := startNth
	for _, w := range wheres {
		if len(w.Values) == 0 {
			continue
		}
		b.WriteString(" AND ")
		b.WriteString(w.Column.Name())
		b.WriteString(" IN (")
		for i, v := range w.Values {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(db.SinglePlaceholder(nth))
			nth++
			vals = append(vals, v)
		}
		b.WriteString(")")
	}
	return b.String(), vals
}

// CompoundWhereOp builds " AND col1 = ? AND col2 > ?" from WhereOp entries.
// Intended to be appended after an existing WHERE clause (e.g. WHERE fk IN (?)).
// startNth is the placeholder numbering offset (e.g. PostgreSQL $n style).
// Returns the SQL fragment and the bind values.
func CompoundWhereOp(wheres []WhereOp, db DB, startNth int) (string, []any) {
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
