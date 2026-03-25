package sqldbs

import "strings"

// WhereClause holds a Cond for building a WHERE clause.
type WhereClause struct {
	Cond Cond
}

// Build produces " WHERE <cond>" with DB-specific placeholder translation.
// startNth is the placeholder numbering offset (number of bind args
// already in the base SQL + 1). Provided by the caller.
func (w WhereClause) Build(db DB, startNth int) (string, []any) {
	if w.Cond == nil {
		return "", nil
	}
	raw, args := w.Cond.BindRepr()
	if raw == "" {
		return "", nil
	}
	var b strings.Builder
	nth := startNth
	for i := 0; i < len(raw); i++ {
		if raw[i] == '?' {
			b.WriteString(db.NthPlaceholder(nth))
			nth++
		} else {
			b.WriteByte(raw[i])
		}
	}
	return " WHERE " + b.String(), args
}
