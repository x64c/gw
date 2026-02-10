package framework

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/x64c/gw/security"
	"github.com/x64c/gw/web/usercookiesession"
)

// PrepareUserCookieSessions prepares UserCookieSessionManager
// Prerequisite: KVDBClient
// Prerequisite: SessionLocks
func (c *Core) PrepareUserCookieSessions() error {
	confFilePath := filepath.Join(c.AppRoot, "config", ".user-cookie-session.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	if c.KVDBClient == nil {
		return errors.New("backend KVDB client not ready")
	}
	if c.SessionLocks == nil {
		return errors.New("sessionlocks not ready")
	}
	mgr := &usercookiesession.Manager{
		AppName:      c.AppName,
		KVDBClient:   c.KVDBClient,
		SessionLocks: c.SessionLocks,
	}
	if err = json.Unmarshal(confBytes, &mgr.Conf); err != nil {
		return err
	}
	// Web Login Session Cipher
	cipher, err := security.NewXChaCha20Poly1305CipherBase64([]byte(mgr.Conf.EncryptionKey))
	if err != nil {
		return fmt.Errorf("NewXChaCha20Poly1305Cipher: %v", err)
	}
	mgr.Cipher = cipher

	c.UserCookieSessionManager = mgr
	return nil
}
