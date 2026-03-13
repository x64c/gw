package framework

import (
	"path/filepath"

	"github.com/x64c/gw/tpl"
)

func (c *Core) PrepareHTMLTemplateStore() error {
	store := tpl.NewHTMLTemplateStore()

	if err := store.LoadFileTemplates(filepath.Join(c.AppRoot, "templates", "html")); err != nil {
		return err
	}
	// ToDo: build derived

	c.HTMLTemplateStore = store
	return nil
}
