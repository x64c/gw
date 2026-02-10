package usercookiesession

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/x64c/gw/db/kvdb"
	"github.com/x64c/gw/security"
)

type Manager struct {
	Conf         Conf
	Cipher       *security.XChaCha20Poly1305Cipher
	AppName      string // for session key, etc.
	SessionLocks *sync.Map
	KVDBClient   kvdb.Client
}

func (m *Manager) SessionIDToKVDBKey(sessionID string) string {
	return m.AppName + ":ucookie_session:" + sessionID
}

func (m *Manager) SessionExistsInKVDB(ctx context.Context, sessionID string) (bool, error) {
	return m.KVDBClient.Exists(ctx, m.SessionIDToKVDBKey(sessionID))
}

func (m *Manager) VerifiySessionCookie(ctx context.Context, r *http.Request) bool {
	sessionCookie, err := r.Cookie(CookieName)
	if err != nil {
		return false
	}
	cookieSessionId, err := m.Cipher.DecodeDecrypt(sessionCookie.Value) // []byte
	if err != nil {
		return false
	}
	found, err := m.SessionExistsInKVDB(ctx, string(cookieSessionId))
	if err != nil {
		return false
	}
	return found
}

func (m *Manager) StoreSessionInKVDB(ctx context.Context, uidStr string) (string, error) {
	cookieSessionID, err := GenerateSessionID()
	if err != nil {
		return "", err
	}
	// Store session_id in KvDB with access_token and refresh_token
	slidingExpiration := time.Duration(m.Conf.ExpireIn) * time.Second
	key := m.SessionIDToKVDBKey(cookieSessionID)
	if err = m.KVDBClient.Set(ctx, key, uidStr, slidingExpiration); err != nil {
		return "", err
	}

	if m.Conf.MaxCntPerUser > 0 {
		usrSessionListKey := fmt.Sprintf("%s:ucookie_sessions:%s", m.AppName, uidStr)
		// SessionList Lock (User Level Lock)
		mu, _ := m.SessionLocks.LoadOrStore(usrSessionListKey, &sync.Mutex{})
		mutex := mu.(*sync.Mutex)

		mutex.Lock() // waits until this gets the lock if it's locked by another goroutine
		defer mutex.Unlock()

		if err = m.KVDBClient.Push(ctx, usrSessionListKey, cookieSessionID); err != nil {
			return "", err
		}

		defer func() {
			_, _ = m.KVDBClient.Expire(
				ctx,
				usrSessionListKey,
				time.Duration(m.Conf.ExpireIn)*time.Second,
			)
		}()

		if err = m.CleanUp(ctx, usrSessionListKey); err != nil {
			return "", err
		}
	}

	return cookieSessionID, nil
}

func (m *Manager) SetSessionCookie(w http.ResponseWriter, sessionID string) error {
	encSessionID, err := m.Cipher.EncryptEncode([]byte(sessionID))
	if err != nil {
		return fmt.Errorf("failed to encrypt cookie session id. %v", err)
	}
	http.SetCookie(w, &http.Cookie{
		Name:  CookieName,
		Value: encSessionID,
		Path:  "/", // Subpaths will get this cookie.
		// Domain: // Cannot be set with `__Host-`
		HttpOnly: true, // JS cannot read it
		Secure:   true, // only sent over HTTPS
		MaxAge:   m.Conf.ExpireIn,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (m *Manager) RefreshSessionCookie(w http.ResponseWriter, encValue string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    encValue, // SAME encrypted value
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   m.Conf.ExpireIn, // reset TTL
		SameSite: http.SameSiteLaxMode,
	})
}

func (m *Manager) RemoveSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Path:     "/",
		MaxAge:   -1, // Delete
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (m *Manager) StoreExternalTokenPairInKVDB(ctx context.Context, sessionID string, apiID string, accessToken string, refreshToken string) error {
	baseKey := m.SessionIDToKVDBKey(sessionID)
	accessTokenKey := baseKey + ":access_tokens"
	refreshTokenKey := baseKey + ":refresh_tokens"

	// If first token pair, set expiration on the containers
	shouldSetExp := false
	found, err := m.KVDBClient.Exists(ctx, accessTokenKey)
	if err != nil || !found {
		shouldSetExp = true
	}

	err = m.KVDBClient.SetField(ctx, accessTokenKey, apiID, accessToken)
	if err != nil {
		return err
	}
	err = m.KVDBClient.SetField(ctx, refreshTokenKey, apiID, refreshToken)
	if err != nil {
		return err
	}

	if shouldSetExp {
		slidingExpiration := time.Duration(m.Conf.ExpireIn) * time.Second
		_, _ = m.KVDBClient.Expire(ctx, accessTokenKey, slidingExpiration)
		_, _ = m.KVDBClient.Expire(ctx, refreshTokenKey, slidingExpiration)
	}

	return nil
}

func (m *Manager) FetchExternalAccessToken(ctx context.Context, sessionID string, apiID string) (string, error) {
	tkn, found, err := m.KVDBClient.GetField(ctx, m.SessionIDToKVDBKey(sessionID)+":access_tokens", apiID)
	if err != nil {
		return "", err
	}
	if !found {
		return "", security.ErrAccessTokenNotFound
	}
	return tkn, nil
}

func (m *Manager) FetchExternalRefreshToken(ctx context.Context, sessionID string, apiID string) (string, error) {
	tkn, found, err := m.KVDBClient.GetField(ctx, m.SessionIDToKVDBKey(sessionID)+":refresh_tokens", apiID)
	if err != nil {
		return "", err
	}
	if !found {
		return "", security.ErrRefreshTokenNotFound
	}
	return tkn, nil
}

func (m *Manager) buildKeysToDel(sessionsToDel []string) []string {
	var keysToDel []string
	if m.Conf.WithExternalTokens {
		keysToDel = make([]string, 0, len(sessionsToDel)*3)
		for _, sid := range sessionsToDel {
			baseKey := m.SessionIDToKVDBKey(sid)
			keysToDel = append(keysToDel,
				baseKey,
				baseKey+":access_tokens",
				baseKey+":refresh_tokens",
			)
		}
		return keysToDel
	}
	keysToDel = make([]string, 0, len(sessionsToDel))
	for _, sid := range sessionsToDel {
		baseKey := m.SessionIDToKVDBKey(sid)
		keysToDel = append(keysToDel, baseKey)
	}
	return keysToDel
}

func (m *Manager) CleanUp(ctx context.Context, usrSessionListKey string) error {
	sessionCnt, err := m.KVDBClient.Len(ctx, usrSessionListKey)
	if err != nil {
		return err
	}
	if sessionCnt <= m.Conf.MaxCntPerUser {
		return nil
	}

	diff := sessionCnt - m.Conf.MaxCntPerUser
	sessionsToDel, err := m.KVDBClient.Range(ctx, usrSessionListKey, 0, diff-1) // []string
	if err != nil {
		return err
	}
	keysToDel := m.buildKeysToDel(sessionsToDel)
	_, _ = m.KVDBClient.Delete(ctx, keysToDel...)
	if err = m.KVDBClient.Trim(ctx, usrSessionListKey, diff, -1); err != nil {
		return err
	}
	return nil
}

func (m *Manager) SessionIDToUIDStrFromKVDB(ctx context.Context, sessionID string) (string, error) {
	return SessionIDToUIDStrFromKVDB(ctx, m, sessionID)
}
