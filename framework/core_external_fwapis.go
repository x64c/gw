package framework

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/x64c/gw/web"
)

func (c *Core) PrepareExternalFWAPIClients(apiIDs ...string) error {
	if c.BaseHttpClient == nil {
		return errors.New("base http client is not ready")
	}
	if len(apiIDs) == 0 {
		return errors.New("no apiID provided")
	}

	c.ExternalFWAPIClients = make(map[string]*ExternalAPIClient, len(apiIDs))

	for _, apiID := range apiIDs {
		confFilePath := filepath.Join(c.AppRoot, "config", fmt.Sprintf(".external-fwapi-%s.json", apiID))
		confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
		if err != nil {
			return err
		}
		fwApiClient := ExternalAPIClient{
			Client: web.ShallowCloneClient(c.BaseHttpClient),
			ApiID:  apiID,
			Core:   c,
		}
		if err = json.Unmarshal(confBytes, &fwApiClient.Conf); err != nil {
			return err
		}
		c.ExternalFWAPIClients[apiID] = &fwApiClient
	}
	return nil
}
