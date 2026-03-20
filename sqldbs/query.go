package sqldbs

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/x64c/gw/orm/coll"
)

// QueryCollectionByColumn queries models where a column matches one or more values.
// Uses WHERE column = ? for single value, WHERE column IN (?, ...) for multiple.
// Returns a collection of scanned models.
func QueryCollectionByColumn[
	M any, // Model struct
	MP ScannableIdentifiable[M, ID], // *Model implementing ScannableIdentifiable[M, ID]
	ID comparable,
	V any,
](
	ctx context.Context,
	db DB,
	sqlSelectBase string,
	column Column,
	values []V,
	orderBys ...OrderBy,
) (*coll.Collection[MP, ID], error) {
	if len(values) == 0 {
		return nil, errors.New("empty values")
	}
	var (
		rows Rows
		err  error
	)
	if len(values) == 1 {
		whereClause := fmt.Sprintf(" WHERE %s = %s", column.Name(), db.SinglePlaceholder())
		sqlStmt := sqlSelectBase + whereClause + OrderByClause(orderBys)
		rows, err = db.QueryRows(ctx, sqlStmt, values[0])
	} else {
		whereClause := fmt.Sprintf(" WHERE %s IN (%s)", column.Name(), db.Placeholders(len(values)))
		sqlStmt := sqlSelectBase + whereClause + OrderByClause(orderBys)
		valuesAsAny := make([]any, len(values))
		for i, v := range values {
			valuesAsAny[i] = v
		}
		rows, err = db.QueryRows(ctx, sqlStmt, valuesAsAny...)
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close() failed: %v", err)
		}
	}()
	return ScanRowsToCollection[M, MP, ID](rows)
}

// QueryCollection queries models into a collection using QueryOpts.
// Builds a standalone clause from QueryOpts, starting with WHERE if any conditions exist.
func QueryCollection[
	M any, // Model struct
	MP ScannableIdentifiable[M, ID], // *Model implementing ScannableIdentifiable[M, ID]
	ID comparable,
](
	ctx context.Context,
	db DB,
	sqlSelectBase string,
	queryOpts QueryOpts,
) (*coll.Collection[MP, ID], error) {
	var args []any
	whereInSQL, whereInArgs := CompoundWhereIn(queryOpts.WhereIns, db, len(args)+1)
	args = append(args, whereInArgs...)
	whereOpSQL, whereOpArgs := CompoundWhereOp(queryOpts.WhereOps, db, len(args)+1)
	args = append(args, whereOpArgs...)
	clause := whereInSQL + whereOpSQL + CompoundWhereNotNullCond(queryOpts.WhereNotNulls) + CompoundWhereNullCond(queryOpts.WhereNulls)
	if queryOpts.HasWhere() {
		// Replace leading " AND" with " WHERE"
		clause = " WHERE" + clause[4:]
	}
	sqlStmt := sqlSelectBase + clause + OrderByClause(queryOpts.OrderBys)
	return RawQueryCollection[M, MP, ID](ctx, db, sqlStmt, args...)
}
