package usercookiesession

import (
	"context"
)

type idCtxKey struct{}

func ContextWithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, idCtxKey{}, sessionID)
}

func SessionIDFromContext(ctx context.Context) (string, bool) {
	ctxVal := ctx.Value(idCtxKey{})
	val, ok := ctxVal.(string)
	return val, ok
}

type uidStrCtxKey struct{}

func ContextWithUIDStr(ctx context.Context, uidStr string) context.Context {
	return context.WithValue(ctx, uidStrCtxKey{}, uidStr)
}

func UIDStrFromContext(ctx context.Context) (string, bool) {
	ctxVal := ctx.Value(uidStrCtxKey{})
	val, ok := ctxVal.(string)
	return val, ok
}
