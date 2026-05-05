package handlerwrappers

import (
	"net/http"

	"github.com/x64c/gw/errs"
	"github.com/x64c/gw/framework"
	"github.com/x64c/gw/security"
	"github.com/x64c/gw/web/responses"
	"github.com/x64c/gw/web/userbearersession"
)

// UserBearerSession validates "Authorization: Bearer ..." against KVDB,
// attaches typed uid + *SessionData[UID] to ctx, and forwards.
type UserBearerSession[UID comparable] struct {
	AppProvider framework.AppProviderFunc
	ParseUID    func(string) (UID, error)
}

func (m *UserBearerSession[UID]) Wrap(inner http.Handler) http.Handler {
	appCore := m.AppProvider().AppCore()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		accessToken := security.ExtractBearerToken(r.Header.Get("Authorization"))
		if accessToken == "" {
			responses.WriteErrorJSON(w, http.StatusUnauthorized, errs.InvalidAccessToken)
			return
		}
		key := appCore.AppName + ":access:" + security.HashHexSHA256(accessToken)
		uidStr, ok, err := appCore.MainKVDB.Get(ctx, key)
		if err != nil {
			responses.WriteErrorJSON(w, http.StatusInternalServerError, errs.KVDB.Wrap(err))
			return
		} else if !ok {
			responses.WriteErrorJSON(w, http.StatusUnauthorized, errs.AccessTokenNotFound)
			return
		}
		uid, err := m.ParseUID(uidStr)
		if err != nil {
			responses.WriteErrorJSON(w, http.StatusInternalServerError, errs.KVDB.WithDetail("parse uid").WithCause(err))
			return
		}
		ctx = userbearersession.WithSessionData(ctx, &userbearersession.SessionData[UID]{UIDStr: uidStr, UID: uid})
		inner.ServeHTTP(w, r.WithContext(ctx))
	})
}
