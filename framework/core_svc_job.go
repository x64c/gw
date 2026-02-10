package framework

import (
	"github.com/x64c/gw/schedjobs"
)

func (c *Core) PrepareJobScheduler() {
	c.JobScheduler = schedjobs.NewScheduler(c.RootCtx)
	c.AddService(c.JobScheduler)
}
