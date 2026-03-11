package framework

import (
	"encoding/json/v2"
	"os"
	"path/filepath"
)

func (c *Core) LoadJwksServiceConf() error {
	confFilePath := filepath.Join(c.AppRoot, "config", ".jwks.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(confBytes, &c.JwksServiceConf); err != nil {
		return err
	}
	return nil
}
