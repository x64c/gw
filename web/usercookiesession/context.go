package usercookiesession

import "context"

type idCtxKey struct{}

func ContextWithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, idCtxKey{}, sessionID)
}

func SessionIDFromContext(ctx context.Context) (string, bool) {
	ctxVal := ctx.Value(idCtxKey{})
	val, ok := ctxVal.(string)
	return val, ok
}

