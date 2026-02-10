package handlerwrappers

import (
	"fmt"
	"net/http"

	"github.com/x64c/gw/authuser"
	"github.com/x64c/gw/contxt"
	"github.com/x64c/gw/framework"
	"github.com/x64c/gw/reason"
	"github.com/x64c/gw/security"
	"github.com/x64c/gw/web/responses"
)

type AuthAPIAccessToken struct {
	AppProvider    framework.AppProviderFunc
	UIDCtxInjector contxt.UnaryInjectorFunc[string] // [optional] ctx with app-specific UID from uidStr
}

// Wrap is a middleware func
// Extracts the Access Token from the request header "Authorization", and Find it in the KVDB.
func (m *AuthAPIAccessToken) Wrap(inner http.Handler) http.Handler {
	if m.UIDCtxInjector == nil {
		// [Default] if omitted, uidStr as UID as-is
		m.UIDCtxInjector = authuser.StrUIDCtxInjector
	}
	appCore := m.AppProvider().AppCore()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Pre-action
		ctx := r.Context()
		// check the access access_token
		accessToken := security.ExtractBearerToken(r.Header.Get("Authorization")) // string
		if accessToken == "" {
			responses.WriteSimpleErrorJSON(w, http.StatusUnauthorized, "access token missing")
			return
		}
		key := appCore.AppName + ":access:" + security.HashHexSHA256(accessToken)
		uidStr, ok, err := appCore.KVDBClient.Get(ctx, key)
		if err != nil {
			responses.WriteSimpleErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("failed to fetch access token info. %v", err))
			return
		} else if !ok {
			responses.WriteErrorJSON(w, http.StatusUnauthorized, reason.AccessTokenExpired, "expired or invalid access token")
			return
		}

		ctx, err = m.UIDCtxInjector(ctx, uidStr)
		if err != nil {
			responses.WriteSimpleErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("failed to parse uid. %v", err))
			return
		}

		// Inner
		inner.ServeHTTP(w, r.WithContext(ctx))

		// Post-action
	})
}
