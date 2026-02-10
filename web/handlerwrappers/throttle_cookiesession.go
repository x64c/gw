package handlerwrappers

import (
	"net/http"
	"time"

	"github.com/x64c/gw/framework"
	"github.com/x64c/gw/web/responses"
	"github.com/x64c/gw/web/usercookiesession"
)

type ThrottleCookieSession struct {
	AppProvider   framework.AppProviderFunc
	BucketGroupID string
}

func (m *ThrottleCookieSession) Wrap(inner http.Handler) http.Handler {
	appCore := m.AppProvider().AppCore()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// Prerequisite _ SessionID
		sessionID, ok := usercookiesession.SessionIDFromContext(ctx)
		if !ok {
			responses.WriteSimpleErrorJSON(w, http.StatusUnauthorized, "invalid session ID")
			return
		}
		// Check Throttle Bucket
		if !appCore.ThrottleBucketStore.Allow(m.BucketGroupID, sessionID, time.Now()) {
			responses.WriteSimpleErrorJSON(w, http.StatusTooManyRequests, "session rate limited")
			return
		}

		// Inner
		inner.ServeHTTP(w, r)

		// Post-action
	})
}
