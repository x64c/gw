package cmdhandlers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/x64c/gw/framework"
	"github.com/x64c/gw/kvdbs"
)

type JwksDeleteOldRsakeys struct {
	AppProvider framework.AppProviderFunc
	KVDB        kvdbs.DB
}

func (*JwksDeleteOldRsakeys) GroupName() string {
	return "jwks"
}

func (h *JwksDeleteOldRsakeys) Command() string {
	return "jwks-delete-old-rsakeys"
}

func (h *JwksDeleteOldRsakeys) Desc() string {
	return "Delete Old RSA key pair files"
}

func (h *JwksDeleteOldRsakeys) Usage() string {
	return h.Command()
}

func (h *JwksDeleteOldRsakeys) HandleCommand(_ []string, w io.Writer) error {
	appCore := h.AppProvider().AppCore()
	// Current Key id
	kid, ok, err := h.KVDB.Get(appCore.RootCtx, appCore.AppName+":kid")
	if err != nil {
		return fmt.Errorf("failed to get current kid: %v\n", err)
	}
	if !ok {
		return fmt.Errorf("kid not found")
	}

	// List files in private key directory
	files, err := os.ReadDir(appCore.JwksServiceConf.PrivateKeyDir)
	if err != nil {
		return fmt.Errorf("failed to read private key dir: %v\n", err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), "_private.pem") && !strings.Contains(f.Name(), kid) {
			// delete old private PEM files
			path := filepath.Join(appCore.JwksServiceConf.PrivateKeyDir, f.Name())
			_ = os.Remove(path)
		}
	}
	// List files in public key directory
	files, err = os.ReadDir(appCore.JwksServiceConf.PublicKeyDir)
	if err != nil {
		return fmt.Errorf("failed to read public key dir: %v\n", err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), "_public.pem") && !strings.Contains(f.Name(), kid) {
			// delete old public PEM files
			path := filepath.Join(appCore.JwksServiceConf.PublicKeyDir, f.Name())
			_ = os.Remove(path)
		}
	}
	_, _ = fmt.Fprintln(w, "Old RSA key pair files deleted")
	return nil
}
