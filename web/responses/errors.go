package responses

// Pre-built errors with default messages.
// For errors that need runtime data in the message, use WithDetail:
//   responses.ErrPermissionDenied.WithDetail("some detail")

var (
	ErrAccessTokenExpired  = &Error{Code: AccessTokenExpired, Message: "expired or invalid access token"}
	ErrRefreshTokenExpired = &Error{Code: RefreshTokenExpired, Message: "refresh token expired"}
	ErrInvalidAccessToken  = &Error{Code: InvalidAccessToken, Message: "invalid access token"}
	ErrInvalidRefreshToken = &Error{Code: InvalidRefreshToken, Message: "invalid refresh token"}

	ErrAPITokenNotFound = &Error{Code: APITokenNotFound, Message: "required API token is missing"}

	ErrCookieSessionExpired          = &Error{Code: CookieSessionExpired, Message: "cookie session expired"}
	ErrCookieSessionAPITokenNotFound = &Error{Code: CookieSessionAPITokenNotFound, Message: "required API token missing on the cookie session"}

	ErrInvalidAuthUID = &Error{Code: InvalidAuthUID, Message: "authenticated user ID missing from context"}

	ErrPermissionDenied = &Error{Code: PermissionDenied, Message: "permission denied"}

	ErrResourceNotFound     = &Error{Code: ResourceNotFound, Message: "resource not found"}
	ErrResourceAccessDenied = &Error{Code: ResourceAccessDenied, Message: "resource access denied"}
)
