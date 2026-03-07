package sqldb

import (
	"context"
	"fmt"

	"github.com/x64c/gw/model"
	"github.com/x64c/gw/nullable"
	"github.com/x64c/gw/orm/coll"
)

// LoadBelongsTo - Load Parents on Children from SQL DB and Link Child-BelongsTo-Parent Relation
// Returns the Parents
func LoadBelongsTo[
	CP model.Identifiable[CID],
	CID comparable,
	P any, // Model struct
	PP ScannableIdentifiable[P, PID],
	PID comparable,
](
	ctx context.Context,
	dbClient Client,
	children *coll.Collection[CP, CID],
	sqlSelectBase string,
	foreignKey func(c CP) PID,
	relationFieldPtr func(c CP) *PP,
) (
	*coll.Collection[PP, PID],
	error,
) {
	fKeysAsAny := coll.CollectUniqueToSlice(children, func(c CP) any { return foreignKey(c) })
	sqlStmt := sqlSelectBase + fmt.Sprintf(" WHERE id IN (%s)", dbClient.Placeholders(len(fKeysAsAny)))
	parents, err := RawQueryCollection[P, PP, PID](ctx, dbClient, sqlStmt, fKeysAsAny...)
	if err != nil {
		return nil, err
	}
	err = coll.LinkBelongsTo[CP, CID, PP, PID](children, parents, foreignKey, relationFieldPtr)
	if err != nil {
		return nil, err
	}
	return parents, nil
}

// LoadOptionalBelongsTo - Load Parents on Children from SQL DB and Link Child-BelongsTo-Parent Relation
// Optional Version: children with null FKs are skipped (their relation field stays nil)
// When accessing the parent model, nil check is required
// Returns the Parent Collection
func LoadOptionalBelongsTo[
	CP model.Identifiable[CID],
	CID comparable,
	P any, // Model struct
	PP ScannableIdentifiable[P, PID],
	PID comparable,
](
	ctx context.Context,
	dbClient Client,
	children *coll.Collection[CP, CID],
	sqlSelectBase string,
	foreignKeyFieldPtr func(c CP) *PID,
	relationFieldPtr func(c CP) *PP,
) (
	*coll.Collection[PP, PID],
	error,
) {
	fKeysAsAny := coll.CollectUniqueToSliceWithSkip(children,
		func(c CP) any {
			ptr := foreignKeyFieldPtr(c)
			if ptr == nil {
				return nil
			}
			return *ptr
		},
		func(v any) bool { return v == nil },
	)
	if len(fKeysAsAny) == 0 {
		return coll.NewEmptyOrderedCollection[PP, PID](), nil
	}
	sqlStmt := sqlSelectBase + fmt.Sprintf(" WHERE id IN (%s)", dbClient.Placeholders(len(fKeysAsAny)))
	parents, err := RawQueryCollection[P, PP, PID](ctx, dbClient, sqlStmt, fKeysAsAny...)
	if err != nil {
		return nil, err
	}
	coll.LinkOptionalBelongsTo[CP, CID, PP, PID](children, parents, foreignKeyFieldPtr, relationFieldPtr)
	return parents, nil
}

// LoadNullableBelongsTo - Convenience wrapper around LoadOptionalBelongsTo for nullable FK fields
// Uses nullable.Nullable[PID] interface to extract the FK pointer via Ptr()
// Returns the Parent Collection
func LoadNullableBelongsTo[
	CP model.Identifiable[CID],
	CID comparable,
	P any, // Model struct
	PP ScannableIdentifiable[P, PID],
	PID comparable,
](
	ctx context.Context,
	dbClient Client,
	children *coll.Collection[CP, CID],
	sqlSelectBase string,
	nullableFKField func(c CP) nullable.Nullable[PID],
	relationFieldPtr func(c CP) *PP,
) (
	*coll.Collection[PP, PID],
	error,
) {
	return LoadOptionalBelongsTo[CP, CID, P, PP, PID](
		ctx, dbClient, children, sqlSelectBase,
		func(c CP) *PID { return nullableFKField(c).Ptr() },
		relationFieldPtr,
	)
}

func LoadHasMany[
	PP model.Identifiable[PID],
	PID comparable,
	C any, // Model struct
	CP ScannableIdentifiable[C, CID],
	CID comparable,
](
	ctx context.Context,
	dbClient Client,
	parents *coll.Collection[PP, PID],
	sqlSelectBase string,
	foreignKeyColumn Column, // on the child
	foreignKey func(CP) PID, // on the child
	relationFieldPtr func(PP) **coll.Collection[CP, CID], // on the parent
	orderBys ...OrderBy,
) (*coll.Collection[CP, CID], error) {
	whereClause := fmt.Sprintf(" WHERE %s IN (%s)", foreignKeyColumn.Name(), dbClient.Placeholders(parents.Len()))
	sqlStmt := sqlSelectBase + whereClause + OrderByClause(orderBys)
	parentIDsAsAny := parents.IDsAsAny()
	children, err := RawQueryCollection[C, CP, CID](ctx, dbClient, sqlStmt, parentIDsAsAny...)
	if err != nil {
		return nil, err
	}
	coll.LinkHasMany[PP, PID, CP, CID](
		parents,
		children,
		foreignKey,
		relationFieldPtr,
	)
	return children, nil
}

// LoadHasManyQueryOpts - Same as LoadHasMany but with QueryOpts for WHERE conditions and ORDER BY.
func LoadHasManyQueryOpts[
	PP model.Identifiable[PID],
	PID comparable,
	C any, // Model struct
	CP ScannableIdentifiable[C, CID],
	CID comparable,
](
	ctx context.Context,
	dbClient Client,
	parents *coll.Collection[PP, PID],
	sqlSelectBase string,
	foreignKeyColumn Column, // on the child
	foreignKey func(CP) PID, // on the child
	relationFieldPtr func(PP) **coll.Collection[CP, CID], // on the parent
	queryOpts QueryOpts,
) (*coll.Collection[CP, CID], error) {
	whereClause := fmt.Sprintf(" WHERE %s IN (%s)", foreignKeyColumn.Name(), dbClient.Placeholders(parents.Len()))
	args := parents.IDsAsAny()
	whereOpExtra, whereOpArgs := CompoundWhereOpCond(queryOpts.WhereOpConds, dbClient, len(args)+1)
	sqlStmt := sqlSelectBase + whereClause + whereOpExtra + CompoundWhereNotNullCond(queryOpts.WhereNotNulls) + CompoundWhereNullCond(queryOpts.WhereNulls) + OrderByClause(queryOpts.OrderBys)
	args = append(args, whereOpArgs...)
	children, err := RawQueryCollection[C, CP, CID](ctx, dbClient, sqlStmt, args...)
	if err != nil {
		return nil, err
	}
	coll.LinkHasMany[PP, PID, CP, CID](
		parents,
		children,
		foreignKey,
		relationFieldPtr,
	)
	return children, nil
}
