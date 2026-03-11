package framework

import "github.com/x64c/gw/tg"

func (c *Core) PrepareTypedGroupRegistry() {
	c.TypedGroupRegistry = make(map[string]tg.RegGrp)
}
