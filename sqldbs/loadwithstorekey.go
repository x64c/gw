package sqldbs

import (
	"context"
	"fmt"

	"github.com/x64c/gw/coll"
	"github.com/x64c/gw/model"
	"github.com/x64c/gw/nullable"
)

// LoadBelongsToWithStoreKey wraps LoadBelongsTo with a RawSQLStore key lookup.
func LoadBelongsToWithStoreKey[
	CP model.Identifiable[CID],
	CID comparable,
	P any,
	PP ScannableIdentifiable[P, PID],
	PID comparable,
](
	ctx context.Context,
	db DB,
	children *coll.Collection[CP, CID],
	storeKey string,
	foreignKey func(c CP) PID,
	relationFieldPtr func(c CP) *PP,
) (
	*coll.Collection[PP, PID],
	error,
) {
	sqlBase, ok := db.MainRawSQLStore().Get(storeKey)
	if !ok {
		return nil, fmt.Errorf("sql statement %q not found in store", storeKey)
	}
	return LoadBelongsTo[CP, CID, P, PP, PID](ctx, db, children, sqlBase, foreignKey, relationFieldPtr)
}

// LoadOptionalBelongsToWithStoreKey wraps LoadOptionalBelongsTo with a RawSQLStore key lookup.
func LoadOptionalBelongsToWithStoreKey[
	CP model.Identifiable[CID],
	CID comparable,
	P any,
	PP ScannableIdentifiable[P, PID],
	PID comparable,
](
	ctx context.Context,
	db DB,
	children *coll.Collection[CP, CID],
	storeKey string,
	foreignKeyFieldPtr func(c CP) *PID,
	relationFieldPtr func(c CP) *PP,
) (
	*coll.Collection[PP, PID],
	error,
) {
	sqlBase, ok := db.MainRawSQLStore().Get(storeKey)
	if !ok {
		return nil, fmt.Errorf("sql statement %q not found in store", storeKey)
	}
	return LoadOptionalBelongsTo[CP, CID, P, PP, PID](ctx, db, children, sqlBase, foreignKeyFieldPtr, relationFieldPtr)
}

// LoadNullableBelongsToWithStoreKey wraps LoadNullableBelongsTo with a RawSQLStore key lookup.
func LoadNullableBelongsToWithStoreKey[
	CP model.Identifiable[CID],
	CID comparable,
	P any,
	PP ScannableIdentifiable[P, PID],
	PID comparable,
](
	ctx context.Context,
	db DB,
	children *coll.Collection[CP, CID],
	storeKey string,
	nullableFKField func(c CP) nullable.Nullable[PID],
	relationFieldPtr func(c CP) *PP,
) (
	*coll.Collection[PP, PID],
	error,
) {
	sqlBase, ok := db.MainRawSQLStore().Get(storeKey)
	if !ok {
		return nil, fmt.Errorf("sql statement %q not found in store", storeKey)
	}
	return LoadNullableBelongsTo[CP, CID, P, PP, PID](ctx, db, children, sqlBase, nullableFKField, relationFieldPtr)
}

// LoadHasManyWithStoreKey wraps LoadHasMany with a RawSQLStore key lookup and FK column name.
func LoadHasManyWithStoreKey[
	PP model.Identifiable[PID],
	PID comparable,
	C any,
	CP ScannableIdentifiable[C, CID],
	CID comparable,
](
	ctx context.Context,
	db DB,
	parents *coll.Collection[PP, PID],
	storeKey string,
	fkColumnName string,
	foreignKey func(CP) PID,
	relationFieldPtr func(PP) **coll.Collection[CP, CID],
	orderBys ...OrderBy,
) (*coll.Collection[CP, CID], error) {
	sqlBase, ok := db.MainRawSQLStore().Get(storeKey)
	if !ok {
		return nil, fmt.Errorf("sql statement %q not found in store", storeKey)
	}
	fkCol, err := NewColumn(fkColumnName)
	if err != nil {
		return nil, fmt.Errorf("invalid foreign key column name %q", fkColumnName)
	}
	return LoadHasMany[PP, PID, C, CP, CID](ctx, db, parents, sqlBase, fkCol, foreignKey, relationFieldPtr, orderBys...)
}

// LoadHasManyQueryOptsWithStoreKey wraps LoadHasManyQueryOpts with a RawSQLStore key lookup and FK column name.
func LoadHasManyQueryOptsWithStoreKey[
	PP model.Identifiable[PID],
	PID comparable,
	C any,
	CP ScannableIdentifiable[C, CID],
	CID comparable,
](
	ctx context.Context,
	db DB,
	parents *coll.Collection[PP, PID],
	storeKey string,
	fkColumnName string,
	foreignKey func(CP) PID,
	relationFieldPtr func(PP) **coll.Collection[CP, CID],
	queryOpts QueryOpts,
) (*coll.Collection[CP, CID], error) {
	sqlBase, ok := db.MainRawSQLStore().Get(storeKey)
	if !ok {
		return nil, fmt.Errorf("sql statement %q not found in store", storeKey)
	}
	fkCol, err := NewColumn(fkColumnName)
	if err != nil {
		return nil, fmt.Errorf("invalid foreign key column name %q", fkColumnName)
	}
	return LoadHasManyQueryOpts[PP, PID, C, CP, CID](ctx, db, parents, sqlBase, fkCol, foreignKey, relationFieldPtr, queryOpts)
}
