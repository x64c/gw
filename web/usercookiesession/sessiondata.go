package usercookiesession

import "context"

// SessionData is the per-request cookie-session payload attached to ctx by
// the auth middleware and read by handlers.
type SessionData[UID comparable] struct {
	ID      string // session ID (KVDB row key)
	UIDStr  string // raw value from KVDB
	UID     UID    // typed (parsed) value
	CSRFTkn string // CSRF token bound to this session
}

func (sd *SessionData[UID]) SessionID() string { return sd.ID }

type ctxKey[UID comparable] struct{}

func WithSessionData[UID comparable](ctx context.Context, sd *SessionData[UID]) context.Context {
	return context.WithValue(ctx, ctxKey[UID]{}, sd)
}

func FromContext[UID comparable](ctx context.Context) (*SessionData[UID], bool) {
	sd, ok := ctx.Value(ctxKey[UID]{}).(*SessionData[UID])
	return sd, ok
}
