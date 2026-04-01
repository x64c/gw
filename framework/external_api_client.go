package framework

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/x64c/gw/reason"
	"github.com/x64c/gw/security"
	"github.com/x64c/gw/web/responses"
	"github.com/x64c/gw/web/usercookiesession"
)

type ExternalAPIClient struct {
	*http.Client // [Embedded]
	ApiID        string
	Conf         *ExternalAPIConf
	Core         *Core
}

func (c *ExternalAPIClient) RequestJWKS(ctx context.Context) (*http.Response, error) {
	upstrUrl := c.Conf.JwksURL
	upstrReq, err := http.NewRequestWithContext(ctx, http.MethodGet, upstrUrl, nil) // *http.Request
	if err != nil {
		return nil, err
	}
	upstrReq.Header.Set("Client-Id", c.Conf.ClientID)
	upstrReq.Header.Set("Content-Type", "application/json")
	upstrReq.Header.Set("Accept", "application/jwk-set+json")
	return http.DefaultClient.Do(upstrReq) // *http.Response
}

// GetJWKS fetches JWKS from .well-known URL for the api
func (c *ExternalAPIClient) GetJWKS(ctx context.Context) (*security.JWKS, error) {
	upstrRes, err := c.RequestJWKS(ctx)
	if err != nil {
		return nil, err
	}
	if upstrRes.StatusCode == http.StatusNotFound {
		return nil, responses.HTTPErrorNotFound
	}
	if upstrRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Status Code: %d", upstrRes.StatusCode)
	}
	defer func() {
		if err = upstrRes.Body.Close(); err != nil {
			log.Printf("[WARN] %v", err)
		}
	}()
	var jwks security.JWKS
	if err = json.UnmarshalRead(upstrRes.Body, &jwks); err != nil {
		return nil, err
	}
	return &jwks, nil
}

func (c *ExternalAPIClient) JWKSFileResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	upstrRes, err := c.RequestJWKS(ctx) // *http.Response
	if err != nil {
		responses.WriteSimpleErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
	if upstrRes.StatusCode == http.StatusNotFound {
		// 404 not found -> raw error message sent before wrapped into JSON
		responses.WriteSimpleErrorJSON(w, http.StatusNotFound, fmt.Sprintf("%v", responses.HTTPErrorNotFound))
		return
	}
	defer func() {
		if closeErr := upstrRes.Body.Close(); closeErr != nil {
			log.Printf("[WARN] %v", closeErr)
		}
	}()
	w.Header().Set("Content-Type", "application/jwk-set+json")
	w.WriteHeader(upstrRes.StatusCode)
	_, err = io.Copy(w, upstrRes.Body)
	if err != nil {
		responses.WriteSimpleErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
}

// RequestJSON sends a request and returns the response.
// The caller is responsible for closing response.Body.
func (c *ExternalAPIClient) RequestJSON(ctx context.Context, accessToken string, method string, endpoint string) (*http.Response, *responses.Error) {
	upstrUrl := c.Conf.Host + endpoint
	upstrReq, err := http.NewRequestWithContext(ctx, method, upstrUrl, nil)
	if err != nil {
		return nil, &responses.Error{Message: err.Error(), Cause: err}
	}

	upstrReq.Header.Set("Client-Id", c.Conf.ClientID)
	upstrReq.Header.Set("Authorization", "Bearer "+accessToken)
	upstrReq.Header.Set("Content-Type", "application/json")
	upstrReq.Header.Set("Accept", "application/json")

	upstrRes, err := c.Do(upstrReq)
	if err != nil {
		return nil, &responses.Error{Message: err.Error(), Cause: err}
	}
	return upstrRes, nil
}

// RequestReissueAccessTokenWithRefreshToken requests the api to reissue access token only with refresh token
func (c *ExternalAPIClient) RequestReissueAccessTokenWithRefreshToken(ctx context.Context, refreshToken string) (*http.Response, error) {
	upstrURL := c.Conf.Host + c.Conf.ReissueAccessTokenEndpoint
	upstrReqBody := security.ReissueAccessTokenRequestBody{
		RefreshToken: refreshToken,
	}
	upstrReqBodyBytes, err := json.Marshal(upstrReqBody)
	if err != nil {
		return nil, err
	}
	upstrReq, err := http.NewRequestWithContext(ctx, http.MethodPost, upstrURL, bytes.NewReader(upstrReqBodyBytes))
	if err != nil {
		return nil, err
	}
	upstrReq.Header.Set("Client-Id", c.Conf.ClientID)
	upstrReq.Header.Set("Content-Type", "application/json")
	upstrReq.Header.Set("Accept", "application/json")
	return c.Do(upstrReq)
}

// RequestReissueAccessTokenWithRefreshTokenAndUserID requests the api to reissue access token with refresh token and user id
func (c *ExternalAPIClient) RequestReissueAccessTokenWithRefreshTokenAndUserID(ctx context.Context, refreshToken string, userIDStr string) (*http.Response, error) {
	upstrURL := c.Conf.Host + c.Conf.ReissueAccessTokenEndpoint
	upstrReqBody := security.ReissueAccessTokenRequestBody{
		RefreshToken: refreshToken,
		UserIDStr:    userIDStr,
	}
	upstrReqBodyBytes, err := json.Marshal(upstrReqBody)
	if err != nil {
		return nil, err
	}
	upstrReq, err := http.NewRequestWithContext(ctx, http.MethodPost, upstrURL, bytes.NewReader(upstrReqBodyBytes))
	if err != nil {
		return nil, err
	}
	upstrReq.Header.Set("Client-Id", c.Conf.ClientID)
	upstrReq.Header.Set("Content-Type", "application/json")
	upstrReq.Header.Set("Accept", "application/json")
	return c.Do(upstrReq)
}

// RefreshAPITokensForCookieSession requests the external API to reissue the access/refresh token pair
// associated with the current cookie session, and updates the KVDB with the new pair.
// A cookie session may have multiple ExternalAPIClients. Call this on each client to refresh its tokens.
func (c *ExternalAPIClient) RefreshAPITokensForCookieSession(ctx context.Context) (*security.AccessTokenPair, int, *responses.Error) {
	sessionID, ok := usercookiesession.SessionIDFromContext(ctx)
	if !ok {
		return nil, http.StatusUnauthorized, &responses.Error{
			Code: reason.CookieSessionExpired, Message: "no session id in the context",
		}
	}
	cookieSessionMgr := c.Core.UserCookieSessionManager

	uidStr, err := cookieSessionMgr.SessionIDToUIDStrFromKVDB(ctx, sessionID)
	if err != nil {
		return nil, http.StatusUnauthorized, &responses.Error{
			Code: reason.CookieSessionExpired, Message: "no session info for the session id",
		}
	}

	refreshToken, err := cookieSessionMgr.FetchExternalRefreshToken(ctx, sessionID, c.ApiID)
	if err != nil {
		return nil, http.StatusUnauthorized, &responses.Error{
			Code: reason.CookieSessionAPITokenNotFound, Message: "no refresh token for the session",
		}
	}

	upstrRes, err := c.RequestReissueAccessTokenWithRefreshTokenAndUserID(ctx, refreshToken, uidStr)
	if err != nil {
		return nil, http.StatusServiceUnavailable, &responses.Error{Message: err.Error()}
	}
	defer func() { _ = upstrRes.Body.Close() }()

	if upstrRes.StatusCode != http.StatusOK {
		// Parse the error response to preserve the logic code
		var resErr responses.Error
		if err = json.UnmarshalRead(upstrRes.Body, &resErr); err != nil {
			return nil, upstrRes.StatusCode, &responses.Error{
				Message: fmt.Sprintf("reissue failed (http %d), could not parse response", upstrRes.StatusCode),
			}
		}
		return nil, upstrRes.StatusCode, &resErr
	}

	var newTokenPair security.AccessTokenPair
	if err = json.UnmarshalRead(upstrRes.Body, &newTokenPair); err != nil {
		return nil, upstrRes.StatusCode, &responses.Error{Message: err.Error()}
	}

	if err = cookieSessionMgr.StoreExternalTokenPairInKVDB(ctx, sessionID, c.ApiID, newTokenPair.AccessToken, newTokenPair.RefreshToken); err != nil {
		return nil, http.StatusInternalServerError, &responses.Error{Message: err.Error()}
	}

	return &newTokenPair, http.StatusOK, nil
}

func (c *ExternalAPIClient) fetchJSON(ctx context.Context, method string, endpoint string) (any, http.Header, int, *responses.Error) { // data, response header, http status, error
	sessionID, ok := usercookiesession.SessionIDFromContext(ctx)
	if !ok {
		return nil, nil, http.StatusUnauthorized, &responses.Error{
			Code: reason.CookieSessionExpired, Message: "no session id in the context",
		}
	}
	accessToken, err := c.Core.UserCookieSessionManager.FetchExternalAccessToken(ctx, sessionID, c.ApiID)
	if err != nil {
		return nil, nil, http.StatusUnauthorized, &responses.Error{
			Code: reason.CookieSessionAPITokenNotFound, Message: fmt.Sprintf("no access token for the api %q in the session", c.ApiID), Cause: err,
		}
	}
	upstrRes, resErr := c.RequestJSON(ctx, accessToken, method, endpoint)
	if resErr != nil {
		return nil, nil, http.StatusServiceUnavailable, resErr
	}
	defer func() { _ = upstrRes.Body.Close() }()

	if upstrRes.StatusCode == http.StatusOK {
		var resData any
		if err = json.UnmarshalRead(upstrRes.Body, &resData); err != nil {
			return nil, nil, http.StatusInternalServerError, &responses.Error{
				Code: reason.JSONUnmarshalFailed, Message: err.Error(), Cause: err,
			}
		}
		return resData, upstrRes.Header, http.StatusOK, nil
	}

	// Error path — parse structured error from API
	var apiErr responses.Error
	if err = json.UnmarshalRead(upstrRes.Body, &apiErr); err != nil {
		return nil, nil, upstrRes.StatusCode, &responses.Error{
			Code: reason.JSONUnmarshalFailed, Message: "failed to unmarshal server error", Cause: err,
		}
	}
	return nil, nil, upstrRes.StatusCode, &apiErr
}

func (c *ExternalAPIClient) FetchJSONRetriable(ctx context.Context, method string, endpoint string) (any, http.Header, int, *responses.Error) { // data, response header, http status, error
	resData, resHeader, httpStatus, resErr := c.fetchJSON(ctx, method, endpoint)
	if resErr == nil {
		return resData, resHeader, httpStatus, nil
	}

	shouldRetry := false
	if resErr.Code != 0 {
		shouldRetry = resErr.Code == reason.AccessTokenExpired
	} else {
		shouldRetry = httpStatus == http.StatusUnauthorized
	}

	if !shouldRetry {
		return nil, nil, httpStatus, resErr
	}

	_, _, refreshErr := c.RefreshAPITokensForCookieSession(ctx)
	if refreshErr != nil {
		return nil, nil, httpStatus, refreshErr
	}
	return c.fetchJSON(ctx, method, endpoint)
}

func (c *ExternalAPIClient) fetchPDFBytes(ctx context.Context, method string, endpoint string) ([]byte, http.Header, int, *responses.Error) { // pdf bytes, response header, http status, error
	sessionID, ok := usercookiesession.SessionIDFromContext(ctx)
	if !ok {
		return nil, nil, http.StatusUnauthorized, &responses.Error{
			Code: reason.CookieSessionExpired, Message: "no session id in the context",
		}
	}
	accessToken, err := c.Core.UserCookieSessionManager.FetchExternalAccessToken(ctx, sessionID, c.ApiID)
	if err != nil {
		return nil, nil, http.StatusUnauthorized, &responses.Error{
			Code: reason.CookieSessionAPITokenNotFound, Message: fmt.Sprintf("no access token for the api %q in the session", c.ApiID), Cause: err,
		}
	}
	upstrUrl := c.Conf.Host + endpoint
	upstrReq, err := http.NewRequestWithContext(ctx, method, upstrUrl, nil)
	if err != nil {
		return nil, nil, http.StatusInternalServerError, &responses.Error{Message: err.Error(), Cause: err}
	}

	upstrReq.Header.Set("Client-Id", c.Conf.ClientID)
	upstrReq.Header.Set("Authorization", "Bearer "+accessToken)
	upstrReq.Header.Set("Content-Type", "application/json")
	upstrReq.Header.Set("Accept", "application/pdf")

	upstrRes, err := c.Do(upstrReq)
	if err != nil {
		return nil, nil, http.StatusServiceUnavailable, &responses.Error{Message: err.Error(), Cause: err}
	}
	defer func() { _ = upstrRes.Body.Close() }()

	if upstrRes.StatusCode == http.StatusOK {
		pdfData, err := io.ReadAll(upstrRes.Body)
		if err != nil {
			return nil, nil, http.StatusInternalServerError, &responses.Error{Message: err.Error(), Cause: err}
		}
		return pdfData, upstrRes.Header, http.StatusOK, nil
	}

	// Error path — parse structured error from API
	var apiErr responses.Error
	if err = json.UnmarshalRead(upstrRes.Body, &apiErr); err != nil {
		return nil, nil, upstrRes.StatusCode, &responses.Error{
			Code: reason.JSONUnmarshalFailed, Message: "failed to unmarshal server error", Cause: err,
		}
	}
	return nil, nil, upstrRes.StatusCode, &apiErr
}

func (c *ExternalAPIClient) FetchPDFBytesRetriable(ctx context.Context, method string, endpoint string) ([]byte, http.Header, int, *responses.Error) { // pdf bytes, response header, http status, error
	pdfData, resHeader, httpStatus, resErr := c.fetchPDFBytes(ctx, method, endpoint)
	if resErr == nil {
		return pdfData, resHeader, httpStatus, nil
	}

	shouldRetry := false
	if resErr.Code != 0 {
		shouldRetry = resErr.Code == reason.AccessTokenExpired
	} else {
		shouldRetry = httpStatus == http.StatusUnauthorized
	}

	if !shouldRetry {
		return nil, nil, httpStatus, resErr
	}

	_, _, refreshErr := c.RefreshAPITokensForCookieSession(ctx)
	if refreshErr != nil {
		return nil, nil, httpStatus, refreshErr
	}
	return c.fetchPDFBytes(ctx, method, endpoint)
}

func (c *ExternalAPIClient) fetchPDFStream(ctx context.Context, method string, endpoint string) (io.ReadCloser, http.Header, int, *responses.Error) { // stream, response header, http status, error
	sessionID, ok := usercookiesession.SessionIDFromContext(ctx)
	if !ok {
		return nil, nil, http.StatusUnauthorized, &responses.Error{
			Code: reason.CookieSessionExpired, Message: "no session id in the context",
		}
	}
	accessToken, err := c.Core.UserCookieSessionManager.FetchExternalAccessToken(ctx, sessionID, c.ApiID)
	if err != nil {
		return nil, nil, http.StatusUnauthorized, &responses.Error{
			Code: reason.CookieSessionAPITokenNotFound, Message: fmt.Sprintf("no access token for the api %q in the session", c.ApiID), Cause: err,
		}
	}
	upstrUrl := c.Conf.Host + endpoint
	upstrReq, err := http.NewRequestWithContext(ctx, method, upstrUrl, nil)
	if err != nil {
		return nil, nil, http.StatusInternalServerError, &responses.Error{Message: err.Error(), Cause: err}
	}

	upstrReq.Header.Set("Client-Id", c.Conf.ClientID)
	upstrReq.Header.Set("Authorization", "Bearer "+accessToken)
	upstrReq.Header.Set("Content-Type", "application/json")
	upstrReq.Header.Set("Accept", "application/pdf")

	upstrRes, err := c.Do(upstrReq)
	if err != nil {
		return nil, nil, http.StatusServiceUnavailable, &responses.Error{Message: err.Error(), Cause: err}
	}

	if upstrRes.StatusCode == http.StatusOK {
		return upstrRes.Body, upstrRes.Header, http.StatusOK, nil
	}

	// Error path — must consume & close body
	defer func() { _ = upstrRes.Body.Close() }()

	var apiErr responses.Error
	if err = json.UnmarshalRead(upstrRes.Body, &apiErr); err != nil {
		return nil, nil, upstrRes.StatusCode, &responses.Error{
			Code: reason.JSONUnmarshalFailed, Message: "failed to unmarshal server error", Cause: err,
		}
	}
	return nil, nil, upstrRes.StatusCode, &apiErr
}

func (c *ExternalAPIClient) FetchPDFStreamRetriable(ctx context.Context, method string, endpoint string) (io.ReadCloser, http.Header, int, *responses.Error) { // stream, response header, http status, error
	stream, resHeader, httpStatus, resErr := c.fetchPDFStream(ctx, method, endpoint)
	if resErr == nil {
		return stream, resHeader, httpStatus, nil
	}

	shouldRetry := false
	if resErr.Code != 0 {
		shouldRetry = resErr.Code == reason.AccessTokenExpired
	} else {
		shouldRetry = httpStatus == http.StatusUnauthorized
	}

	if !shouldRetry {
		return nil, nil, httpStatus, resErr
	}

	_, _, refreshErr := c.RefreshAPITokensForCookieSession(ctx)
	if refreshErr != nil {
		return nil, nil, httpStatus, refreshErr
	}
	return c.fetchPDFStream(ctx, method, endpoint)
}
