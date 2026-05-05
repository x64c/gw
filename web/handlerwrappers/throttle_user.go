package handlerwrappers

import (
	"net/http"
	"time"

	"github.com/x64c/gw/errs"
	"github.com/x64c/gw/framework"
	"github.com/x64c/gw/web/responses"
)

// throttleUser checks the user-keyed throttle bucket and writes a 429 if rate-limited.
// Returns true if the request may proceed.
// Used by ThrottleBearerUser and ThrottleCookieUser to share the bucket-check logic.
func throttleUser(w http.ResponseWriter, appCore *framework.Core, uidStr, bucketGroupID string) bool {
	if !appCore.ThrottleBucketStore.Allow(bucketGroupID, uidStr, time.Now()) {
		responses.WriteErrorJSON(w, http.StatusTooManyRequests, errs.RateLimited)
		return false
	}
	return true
}
