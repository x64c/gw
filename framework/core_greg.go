package framework

import "github.com/x64c/gw/greg"

func (c *Core) PrepareTypedGroupRegistry() {
	c.TypedGroupRegistry = make(map[string]greg.RegGrp)
}
