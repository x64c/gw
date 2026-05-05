package errs

// Pre-built errors with default messages.
// For errors that need runtime data in the message, use WithDetail:
//   errs.PermissionDenied.WithDetail("some detail")
//
// Framework-level logic error codes (1000-1999).
// App-level codes use 2000+ in their own `reasons` package.
// Code 0 = no logic code (client falls back to HTTP status).

var (
	// Bearer Session (Access/Refresh Tokens)

	AccessTokenNotFound  = &Error{Name: "AccessTokenNotFound", Code: 1000, Message: "access token not found"}
	RefreshTokenNotFound = &Error{Name: "RefreshTokenNotFound", Code: 1001, Message: "refresh token not found"}
	InvalidAccessToken   = &Error{Name: "InvalidAccessToken", Code: 1002, Message: "invalid access token"}
	InvalidRefreshToken  = &Error{Name: "InvalidRefreshToken", Code: 1003, Message: "invalid refresh token"}

	// Cookie Session

	CookieNotFound        = &Error{Name: "CookieNotFound", Code: 1100, Message: "cookie not found"}
	InvalidCookie         = &Error{Name: "InvalidCookie", Code: 1101, Message: "invalid cookie"}
	CookieSessionNotFound = &Error{Name: "CookieSessionNotFound", Code: 1102, Message: "cookie session not found"}
	CSRFTokenNotFound     = &Error{Name: "CSRFTokenNotFound", Code: 1110, Message: "CSRF token not found"}
	InvalidCSRFToken      = &Error{Name: "InvalidCSRFToken", Code: 1111, Message: "invalid CSRF token"}

	// OAuth

	InvalidOAuthState = &Error{Name: "InvalidOAuthState", Code: 1200, Message: "invalid OAuth state"}

	// Data Format & Serialization

	JSONMarshalFailed   = &Error{Name: "JSONMarshalFailed", Code: 1300, Message: "failed to marshal JSON"}
	JSONUnmarshalFailed = &Error{Name: "JSONUnmarshalFailed", Code: 1301, Message: "failed to unmarshal JSON"}

	// Access Control (Permissions, Resources, Throttling)

	DataMissingInContext = &Error{Name: "DataMissingInContext", Code: 1400, Message: "data missing in context"} // expected ctx attachment absent (middleware misconfiguration / bypass)
	InvalidAuthUID       = &Error{Name: "InvalidAuthUID", Code: 1401, Message: "authenticated user ID missing from context"}
	PermissionDenied     = &Error{Name: "PermissionDenied", Code: 1410, Message: "permission denied"}           // user lacks required permission
	ResourceNotFound     = &Error{Name: "ResourceNotFound", Code: 1420, Message: "resource not found"}          // expected resource must exist but is missing
	ResourceAccessDenied = &Error{Name: "ResourceAccessDenied", Code: 1421, Message: "resource access denied"}  // resource exists but user cannot access it
	ResourceUnavailable  = &Error{Name: "ResourceUnavailable", Code: 1422, Message: "resource unavailable"}     // resource exists but is not currently available (temporarily or permanently)
	RateLimited          = &Error{Name: "RateLimited", Code: 1430, Message: "rate limited"}                     // request throttled (per-user / per-session / per-IP bucket exceeded)

	// DB

	KVDB               = &Error{Name: "KVDB", Code: 1600, Message: "kvdb error"}                              // general key-value store error
	SQLDB              = &Error{Name: "SQLDB", Code: 1610, Message: "sql db error"}                            // general SQL/database error
	SQLNotFoundInStore = &Error{Name: "SQLNotFoundInStore", Code: 1611, Message: "sql statement not found in store"} // SQL statement not found in RawSQLStore

	// Relation

	RelBelongsToLinkFailed = &Error{Name: "RelBelongsToLinkFailed", Code: 1700, Message: "relation BelongsTo link failed"} // parent not found for child's FK during LinkBelongsTo

	// Upstream

	UpstreamAccessTokenNotFound  = &Error{Name: "UpstreamAccessTokenNotFound", Code: 1800, Message: "upstream access token not found"}   // access token missing to authenticate with an upstream server
	UpstreamRefreshTokenNotFound = &Error{Name: "UpstreamRefreshTokenNotFound", Code: 1801, Message: "upstream refresh token not found"} // refresh token missing to reissue an upstream access token
	InvalidUpstreamAccessToken   = &Error{Name: "InvalidUpstreamAccessToken", Code: 1802, Message: "invalid upstream access token"}
	InvalidUpstreamRefreshToken  = &Error{Name: "InvalidUpstreamRefreshToken", Code: 1803, Message: "invalid upstream refresh token"}
	Upstream                     = &Error{Name: "Upstream", Code: 1810, Message: "upstream error"} // failure during upstream interaction (build/transport/server)

	// Misc

	PDFBuildFailed = &Error{Name: "PDFBuildFailed", Code: 1900, Message: "failed to build PDF"}
)
