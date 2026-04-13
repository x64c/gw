package errs

// Pre-built errors with default messages.
// For errors that need runtime data in the message, use WithDetail:
//   errs.PermissionDenied.WithDetail("some detail")
//
// Framework-level logic error codes (1000-1999).
// App-level codes use 2000+ in their own `reasons` package.
// Code 0 = no logic code (client falls back to HTTP status).

var (
	// Access Tokens

	AccessTokenExpired  = &Error{Name: "AccessTokenExpired", Code: 1000, Message: "expired or invalid access token"}  // access token expired or not found in store
	RefreshTokenExpired = &Error{Name: "RefreshTokenExpired", Code: 1001, Message: "refresh token expired"}            // refresh token expired, used, or not found in store
	InvalidAccessToken  = &Error{Name: "InvalidAccessToken", Code: 1002, Message: "invalid access token"}              // access token exists but fails validation
	InvalidRefreshToken = &Error{Name: "InvalidRefreshToken", Code: 1003, Message: "invalid refresh token"}            // refresh token exists but uid/client mismatch

	// Cookie Session

	CookieSessionNotFound         = &Error{Name: "CookieSessionNotFound", Code: 1100, Message: "cookie session not found"}                          // cookie session not found in context or store (expired)
	CookieSessionAPITokenNotFound = &Error{Name: "CookieSessionAPITokenNotFound", Code: 1101, Message: "required API token missing on the cookie session"} // cookie session alive but required API token missing

	// Data Format & Serialization

	JSONMarshalFailed   = &Error{Name: "JSONMarshalFailed", Code: 1200, Message: "failed to marshal JSON"}
	JSONUnmarshalFailed = &Error{Name: "JSONUnmarshalFailed", Code: 1201, Message: "failed to unmarshal JSON"}

	// Access Control (Permissions & Resources)

	InvalidAuthUID       = &Error{Name: "InvalidAuthUID", Code: 1300, Message: "authenticated user ID missing from context"}
	PermissionDenied     = &Error{Name: "PermissionDenied", Code: 1310, Message: "permission denied"}           // user lacks required permission
	ResourceNotFound     = &Error{Name: "ResourceNotFound", Code: 1320, Message: "resource not found"}          // expected resource must exist but is missing
	ResourceAccessDenied = &Error{Name: "ResourceAccessDenied", Code: 1321, Message: "resource access denied"}  // resource exists but user cannot access it
	ResourceUnavailable  = &Error{Name: "ResourceUnavailable", Code: 1322, Message: "resource unavailable"}     // resource exists but is currently not available

	// DB

	KVDB               = &Error{Name: "KVDB", Code: 1700, Message: "kvdb error"}                              // general key-value store error
	SQLDB              = &Error{Name: "SQLDB", Code: 1710, Message: "sql db error"}                            // general SQL/database error
	SQLNotFoundInStore = &Error{Name: "SQLNotFoundInStore", Code: 1711, Message: "sql statement not found in store"} // SQL statement not found in RawSQLStore

	// Relation

	RelBelongsToLinkFailed = &Error{Name: "RelBelongsToLinkFailed", Code: 1800, Message: "relation BelongsTo link failed"} // parent not found for child's FK during LinkBelongsTo

	// Upstream

	APITokenNotFound = &Error{Name: "APITokenNotFound", Code: 1900, Message: "required API token is missing"} // token missing to authenticate with an upstream server
	Upstream         = &Error{Name: "Upstream", Code: 1901, Message: "upstream error"}                         // failure during upstream interaction (build/transport/server)
)
