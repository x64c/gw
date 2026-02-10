package framework

import (
	"time"

	"github.com/x64c/gw/throttle"
)

func (c *Core) PrepareThrottleBucketStore(cleanupCycle time.Duration, cleanupOlderThan time.Duration) {
	c.ThrottleBucketStore = throttle.NewBucketStore(c.RootCtx, cleanupCycle, cleanupOlderThan)
	c.AddService(c.ThrottleBucketStore)
}
