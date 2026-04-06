package responses

// Framework-level logic error codes (1000-1999).
// App-level codes use 2000+ in their own `reasons` package.
// Code 0 = no logic code (client falls back to HTTP status).
const (
	// Token validity

	AccessTokenExpired  = 1000 // access token expired or not found in store
	RefreshTokenExpired = 1001 // refresh token expired, used, or not found in store
	InvalidAccessToken  = 1002 // access token exists but fails validation
	InvalidRefreshToken = 1003 // refresh token exists but uid/client mismatch

	// Token availability

	APITokenNotFound = 1100 // required API token is missing

	// Cookie Session

	CookieSessionExpired          = 1200 // cookie session not found or expired in store
	CookieSessionAPITokenNotFound = 1201 // cookie session alive but required API token missing

	// Auth

	InvalidAuthUID = 1300 // authenticated user ID missing from context

	// Permission

	PermissionDenied = 1400 // user lacks required permission

	// Resource

	ResourceNotFound     = 1500 // expected resource must exist but is missing
	ResourceAccessDenied = 1501 // resource exists but user cannot access it

	// JSON

	JSONUnmarshalFailed = 1600 // failed to unmarshal JSON response

	// SQL

	SQLError           = 1700 // general SQL/database error
	SQLNotFoundInStore = 1710 // SQL statement not found in RawSQLStore
)
