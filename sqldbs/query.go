package sqldbs

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/x64c/gw/orm/coll"
)

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
