# Response Error Guide

## HTTP Status + Logic Code Combinations

The API returns `{"message": "...", "code": <int>}` on errors. The `code` field is the logic code — clients check it first, fall back to HTTP status if code is 0.

### HTTP 401 Unauthorized — "Who are you?"

| Logic Code | Meaning | Client Action |
|------------|---------|---------------|
| `AccessTokenExpired` | Access token expired or not in store | Reissue with refresh token, retry |
| `RefreshTokenExpired` | Refresh token dead | Must re-login |
| `InvalidAccessToken` | Access token fails validation | Must re-login |
| `InvalidRefreshToken` | Refresh token fails validation | Must re-login |
| 0 (no code) | Generic auth failure | Try reissue, if fails re-login |

### HTTP 403 Forbidden — "I know who you are, but no"

| Logic Code | Meaning | Client Action |
|------------|---------|---------------|
| `PermissionDenied` | User lacks required permission (action-level) | Show error, do NOT retry |
| `ResourceAccessDenied` | Resource exists but user cannot access it (data-level) | Show error, do NOT retry |
| 0 (no code) | Generic forbidden | Show error |

### HTTP 404 Not Found

| Response | Meaning |
|----------|---------|
| Raw (no JSON body) | URL/endpoint does not exist |
| `ResourceNotFound` | Endpoint hit, but required data/resource is missing |

### HTTP 500 Internal Server Error

| Logic Code | Meaning |
|------------|---------|
| `InvalidAuthUID` | Auth user ID missing from context |
| `JSONUnmarshalFailed` | Failed to parse JSON |
| `SQLNotFoundInStore` | SQL statement missing from store |
| 0 (no code) | Generic server error |

## Guidelines

- **Use named constants only** (e.g. `reason.ResourceNotFound`), never raw integers. The underlying numbers are internal (because Go has no enum) and may change between versions.
- **API responses to clients**: use logic code for detail.
- **Low-level, internal endpoints**: simple error messages can be enough.
- **Resource not found vs empty data**: Use `ResourceNotFound` error only when the specific resource MUST exist but is missing.
- **Permission vs Resource access**: `PermissionDenied` = action-level (can't do this operation). `ResourceAccessDenied` = data-level (can't access this specific item).
- **App-specific codes**: define your own in a separate package. Framework codes are below 2000, app codes start at 2000+.
