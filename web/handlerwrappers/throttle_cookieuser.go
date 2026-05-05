package handlerwrappers

import (
	"net/http"

	"github.com/x64c/gw/errs"
	"github.com/x64c/gw/framework"
	"github.com/x64c/gw/web/responses"
	"github.com/x64c/gw/web/usercookiesession"
)

// ThrottleCookieUser throttles by uid extracted from the user-cookie-session in ctx.
type ThrottleCookieUser[UID comparable] struct {
	AppProvider   framework.AppProviderFunc
	BucketGroupID string
}

func (m *ThrottleCookieUser[UID]) Wrap(inner http.Handler) http.Handler {
	appCore := m.AppProvider().AppCore()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sd, ok := usercookiesession.FromContext[UID](r.Context())
		if !ok {
			responses.WriteErrorJSON(w, http.StatusInternalServerError, errs.DataMissingInContext.WithDetail("SessionData"))
			return
		}
		if !throttleUser(w, appCore, sd.UIDStr, m.BucketGroupID) {
			return
		}
		inner.ServeHTTP(w, r)
	})
}
