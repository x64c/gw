package reason

// Framework-level logic error codes (1000-1999).
// App-level codes use 2000+ in their own `reasons` package.
// Code 0 = no logic code (client falls back to HTTP status).
const (
	// Token — validity (1000-1049)

	AccessTokenExpired  = 1000 // access token expired or not found in store
	RefreshTokenExpired = 1001 // refresh token expired, used, or not found in store
	InvalidAccessToken  = 1002 // access token exists but fails validation
	InvalidRefreshToken = 1003 // refresh token exists but uid/client mismatch

	// Token — availability (1050-1099)

	APITokenNotFound = 1050 // required API token is missing

	// Cookie Session (1100-1199)

	CookieSessionExpired          = 1100 // cookie session not found or expired in store
	CookieSessionAPITokenNotFound = 1101 // cookie session alive but required API token missing

	// Infrastructure (1200-1299)

	JSONUnmarshalFailed = 1200 // failed to unmarshal JSON response
)
