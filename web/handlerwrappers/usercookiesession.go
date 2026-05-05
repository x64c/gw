package handlerwrappers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/x64c/gw/errs"
	"github.com/x64c/gw/framework"
	"github.com/x64c/gw/kvdbs"
	"github.com/x64c/gw/web/responses"
	"github.com/x64c/gw/web/usercookiesession"
)

type UserCookieSession[UID comparable] struct {
	AppProvider framework.AppProviderFunc
	ParseUID    func(string) (UID, error)
}

func (m *UserCookieSession[UID]) Wrap(inner http.Handler) http.Handler {
	appCore := m.AppProvider().AppCore()
	cookieSessionMgr := appCore.UserCookieSessionManager
	switch cookieSessionMgr.Conf.ExpireMode {
	case usercookiesession.ExpireAbsolute:
		return m.absoluteExpHandler(inner, cookieSessionMgr)
	case usercookiesession.ExpireSliding:
		return m.slidingExpHandler(inner, cookieSessionMgr)
	default:
		log.Fatal("[ERROR] invalid cookie session expiration mode")
		return nil
	}
}

func (m *UserCookieSession[UID]) authenticateCookieSession(
	w http.ResponseWriter, r *http.Request, cookieSessionMgr *usercookiesession.Manager,
) (
	ctx context.Context, sessionCookie *http.Cookie, sessionID string, uidStr string, ok bool, // ok to proceed
) {
	ctx = r.Context()
	// If Logged-in, Session Cookie must be shipped in the request
	sessionCookie, err := r.Cookie(usercookiesession.CookieName)
	if err != nil { // http.ErrNoCookie
		// Session Cookie Not Found = Non-login Hit to Auth-protected Endpoints
		// Redirect to Login page setting Intended URI Cookie
		// ToDo: flash msg "Login Required"
		usercookiesession.SetIntendedURICookie(w, r, 60) // short-lived cookie
		http.Redirect(w, r, cookieSessionMgr.Conf.LoginPath+"?endpoint=protected", http.StatusSeeOther)
		return nil, nil, "", "", false
	}
	sessionIDBytes, err := cookieSessionMgr.Cipher.DecodeDecrypt(sessionCookie.Value)
	if err != nil {
		responses.WriteSimpleErrorJSON(w, http.StatusUnauthorized, fmt.Sprintf("invalid session. %v", err))
		return nil, nil, "", "", false
	}
	sessionID = string(sessionIDBytes)

	key := cookieSessionMgr.SessionIDToKVDBKey(sessionID)
	fields, err := cookieSessionMgr.KVDB.GetFields(ctx, key, "uid", "csrf")
	if err != nil {
		responses.WriteSimpleErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("failed to check session. %v", err))
		return nil, nil, "", "", false
	}
	uidStr, hasUID := fields["uid"]
	if !hasUID {
		// Session Not Found. Session might have been Expired.
		// Redirect to Login page Clearing Session Cookie
		cookieSessionMgr.RemoveSessionCookie(w)
		usercookiesession.SetIntendedURICookie(w, r, 60)
		http.Redirect(w, r, cookieSessionMgr.Conf.LoginPath+"?session=expired", http.StatusSeeOther) // ToDo: abstract this.
		return nil, nil, "", "", false
	}
	csrfTkn := fields["csrf"]

	uid, err := m.ParseUID(uidStr)
	if err != nil {
		responses.WriteErrorJSON(w, http.StatusInternalServerError, errs.KVDB.WithDetail("parse uid").WithCause(err))
		return
	}

	// new context for the next handler
	ctx = usercookiesession.ContextWithSessionID(ctx, sessionID) // legacy
	ctx = usercookiesession.WithSessionData(ctx, &usercookiesession.SessionData[UID]{
		ID:      sessionID,
		UIDStr:  uidStr,
		UID:     uid,
		CSRFTkn: csrfTkn,
	})
	return ctx, sessionCookie, sessionID, uidStr, true
}

func (m *UserCookieSession[UID]) absoluteExpHandler(inner http.Handler, cookieSessionMgr *usercookiesession.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, _, _, _, ok := m.authenticateCookieSession(w, r, cookieSessionMgr)
		if !ok {
			return
		}

		// Inner
		inner.ServeHTTP(w, r.WithContext(ctx))

		// Post-action
	})
}

func (m *UserCookieSession[UID]) slidingExpHandler(inner http.Handler, cookieSessionMgr *usercookiesession.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, sessionCookie, sessionID, uidStr, ok := m.authenticateCookieSession(w, r, cookieSessionMgr)
		if !ok {
			return
		}

		// Sliding-specific
		baseKey := cookieSessionMgr.SessionIDToKVDBKey(sessionID)
		ttl, state, err := cookieSessionMgr.KVDB.TTL(ctx, baseKey)
		if err == nil && state == kvdbs.TTLExpiring && ttl < time.Duration(cookieSessionMgr.Conf.ExtendThreshold)*time.Second {
			slidingExpiration := time.Duration(cookieSessionMgr.Conf.ExpireIn) * time.Second
			_, _ = cookieSessionMgr.KVDB.Expire(ctx, baseKey, slidingExpiration)
			if cookieSessionMgr.Conf.WithExternalTokens {
				_, _ = cookieSessionMgr.KVDB.Expire(ctx, baseKey+":access_tokens", slidingExpiration)
				_, _ = cookieSessionMgr.KVDB.Expire(ctx, baseKey+":refresh_tokens", slidingExpiration)
			}
			if cookieSessionMgr.Conf.MaxCntPerUser > 0 {
				usrSessionListKey := fmt.Sprintf("%s:ucookie_sessions:%s", cookieSessionMgr.AppName, uidStr)
				_, _ = cookieSessionMgr.KVDB.Expire(ctx, usrSessionListKey, slidingExpiration)
			}
			encSessionID := sessionCookie.Value
			cookieSessionMgr.RefreshSessionCookie(w, encSessionID)
		}

		// Inner
		inner.ServeHTTP(w, r.WithContext(ctx))

		// Post-action
	})
}
