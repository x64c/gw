package web

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/x64c/gw/errs"
	"github.com/x64c/gw/security"
	"github.com/x64c/gw/web/responses"
	"github.com/x64c/gw/web/usercookiesession"
)

type ExternalFWAPIClient struct {
	*http.Client                                       // [Embedded]
	ApiID                    string
	Conf                     *ExternalFWAPIConf
	UserCookieSessionManager *usercookiesession.Manager
}

func (c *ExternalFWAPIClient) RequestJWKS(ctx context.Context) (*http.Response, error) {
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
func (c *ExternalFWAPIClient) GetJWKS(ctx context.Context) (*security.JWKS, error) {
	upstrRes, err := c.RequestJWKS(ctx)
	if err != nil {
		return nil, err
	}
	if upstrRes.StatusCode == http.StatusNotFound {
		return nil, errors.New("JWKS not found")
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

func (c *ExternalFWAPIClient) JWKSFileResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	upstrRes, err := c.RequestJWKS(ctx) // *http.Response
	if err != nil {
		responses.WriteSimpleErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
	if upstrRes.StatusCode == http.StatusNotFound {
		// 404 not found -> raw error message sent before wrapped into JSON
		responses.WriteSimpleErrorJSON(w, http.StatusNotFound, "JWKS not found")
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

// RequestReissueAccessTokenWithRefreshToken requests the api to reissue access token only with refresh token
func (c *ExternalFWAPIClient) RequestReissueAccessTokenWithRefreshToken(ctx context.Context, refreshToken string) (*http.Response, error) {
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
func (c *ExternalFWAPIClient) RequestReissueAccessTokenWithRefreshTokenAndUserID(ctx context.Context, refreshToken string, userIDStr string) (*http.Response, error) {
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
// A cookie session may have multiple ExternalFWAPIClients. Call this on each client to refresh its tokens.
func (c *ExternalFWAPIClient) RefreshAPITokensForCookieSession(ctx context.Context) (*security.AccessTokenPair, int, *errs.Error) {
	sessionID, ok := usercookiesession.SessionIDFromContext(ctx)
	if !ok {
		return nil, http.StatusUnauthorized, errs.CookieSessionNotFound.WithDetail("no session id in the context")
	}
	cookieSessionMgr := c.UserCookieSessionManager

	uidStr, err := cookieSessionMgr.SessionIDToUIDStrFromKVDB(ctx, sessionID)
	if err != nil {
		return nil, http.StatusUnauthorized, errs.CookieSessionNotFound.WithDetail("no session info for the session id")
	}

	refreshToken, err := cookieSessionMgr.FetchExternalRefreshToken(ctx, sessionID, c.ApiID)
	if err != nil {
		return nil, http.StatusUnauthorized, errs.UpstreamRefreshTokenNotFound.WithDetail("no refresh token")
	}

	upstrRes, err := c.RequestReissueAccessTokenWithRefreshTokenAndUserID(ctx, refreshToken, uidStr)
	if err != nil {
		return nil, http.StatusServiceUnavailable, errs.Upstream.WithDetail(err.Error())
	}
	defer func() { _ = upstrRes.Body.Close() }()

	if upstrRes.StatusCode != http.StatusOK {
		// Parse the error response to preserve the logic code
		var resErr errs.Error
		if err = json.UnmarshalRead(upstrRes.Body, &resErr); err != nil {
			return nil, upstrRes.StatusCode, errs.JSONUnmarshalFailed.WithDetail(fmt.Sprintf("reissue failed (http %d), could not parse response", upstrRes.StatusCode))
		}
		return nil, upstrRes.StatusCode, &resErr
	}

	var newTokenPair security.AccessTokenPair
	if err = json.UnmarshalRead(upstrRes.Body, &newTokenPair); err != nil {
		return nil, upstrRes.StatusCode, errs.JSONUnmarshalFailed.WithDetail(err.Error())
	}

	if err = cookieSessionMgr.StoreExternalTokenPairInKVDB(ctx, sessionID, c.ApiID, newTokenPair.AccessToken, newTokenPair.RefreshToken); err != nil {
		return nil, http.StatusInternalServerError, errs.KVDB.WithDetail(err.Error())
	}

	return &newTokenPair, http.StatusOK, nil
}

// doRequest performs the common request flow for ExternalFWAPIClient fetch methods:
// session lookup, access token fetch, request build, send, and error-status parsing.
//
// payload (optional) carries extra headers and a Request Body provider closure. Pass
// nil for a body-less Request with no extra headers. The closure is called fresh on
// every invocation (so framework-level retries get a new reader each time) and is
// also wired to stdlib's Request.GetBody so stdlib's own replay paths (HTTP redirects,
// HTTP/2 retries) can rebuild the Request Body too.
//
// On success (HTTP 200), returns *http.Response with the Response Body NOT consumed —
// caller is responsible for closing it. On any failure, returns nil response and the parsed errs.
func (c *ExternalFWAPIClient) doRequest(ctx context.Context, method, endpoint string, payload *RequestPayload) (*http.Response, int, *errs.Error) {
	sessionID, ok := usercookiesession.SessionIDFromContext(ctx)
	if !ok {
		return nil, http.StatusUnauthorized, errs.CookieSessionNotFound.WithDetail("no session id in the context")
	}
	accessToken, err := c.UserCookieSessionManager.FetchExternalAccessToken(ctx, sessionID, c.ApiID)
	if err != nil {
		return nil, http.StatusUnauthorized, errs.UpstreamAccessTokenNotFound.WithDetail(fmt.Sprintf("no access token for the api %q in the session", c.ApiID)).WithCause(err)
	}

	// Build the Request Body fresh from the caller's BodyProvider (if any).
	var reqBodyReader io.Reader
	if payload != nil && payload.BodyProvider != nil {
		reqBodyReader, err = payload.BodyProvider()
		if err != nil {
			return nil, http.StatusBadRequest, errs.Upstream.Wrap(err)
		}
	}

	upstrUrl := c.Conf.Host + endpoint
	upstrReq, err := http.NewRequestWithContext(ctx, method, upstrUrl, reqBodyReader)
	if err != nil {
		return nil, http.StatusBadRequest, errs.Upstream.Wrap(err)
	}

	// Headers in three layers, in order:
	//   1. Framework defaults (Content-Type = JSON, etc.) — caller can override below
	//   2. Caller's headers from payload — overwrites defaults if set
	//   3. Framework auth headers — set last so they always win (caller can't override)
	upstrReq.Header.Set("Content-Type", "application/json") // framework default; caller may override via payload.Headers
	if payload != nil {
		for k, vs := range payload.Headers {
			upstrReq.Header[k] = vs
		}
	}
	upstrReq.Header.Set("Client-Id", c.Conf.ClientID)
	upstrReq.Header.Set("Authorization", "Bearer "+accessToken)

	// Wire stdlib's GetBody from the caller's BodyProvider so stdlib's replays
	// (HTTP redirects, HTTP/2 retries) also rebuild the Request Body fresh.
	if payload != nil && payload.BodyProvider != nil {
		upstrReq.GetBody = func() (io.ReadCloser, error) {
			r, err := payload.BodyProvider()
			if err != nil {
				return nil, err
			}
			return io.NopCloser(r), nil
		}
	}

	upstrRes, err := c.Do(upstrReq)
	if err != nil {
		return nil, http.StatusBadGateway, errs.Upstream.Wrap(err)
	}

	if upstrRes.StatusCode != http.StatusOK {
		// Error path — parse structured error from API, then close the Response Body
		defer func() { _ = upstrRes.Body.Close() }()
		var apiErr errs.Error
		if err = json.UnmarshalRead(upstrRes.Body, &apiErr); err != nil {
			return nil, upstrRes.StatusCode, errs.JSONUnmarshalFailed.WithDetail("failed to unmarshal server error").WithCause(err)
		}
		return nil, upstrRes.StatusCode, &apiErr
	}

	return upstrRes, http.StatusOK, nil
}

func (c *ExternalFWAPIClient) fetchJSON(ctx context.Context, method string, endpoint string, reqPayload *RequestPayload) (any, http.Header, int, *errs.Error) { // data, response header, http status, error
	// Build the actual payload: caller's body + caller's headers + framework's Accept (forced).
	// Accept is fetchJSON's contract — caller can't override it.
	actualReqPayload := &RequestPayload{Headers: http.Header{}}
	if reqPayload != nil {
		for k, vs := range reqPayload.Headers {
			actualReqPayload.Headers[k] = vs
		}
		actualReqPayload.BodyProvider = reqPayload.BodyProvider
	}
	actualReqPayload.Headers.Set("Accept", "application/json")

	upstrRes, status, resErr := c.doRequest(ctx, method, endpoint, actualReqPayload)
	if resErr != nil {
		return nil, nil, status, resErr
	}
	defer func() { _ = upstrRes.Body.Close() }()

	var resData any
	if err := json.UnmarshalRead(upstrRes.Body, &resData); err != nil {
		return nil, nil, http.StatusInternalServerError, errs.JSONUnmarshalFailed.Wrap(err)
	}
	return resData, upstrRes.Header, http.StatusOK, nil
}

func (c *ExternalFWAPIClient) FetchJSONRetriable(ctx context.Context, method string, endpoint string, reqPayload *RequestPayload) (any, http.Header, int, *errs.Error) { // data, response header, http status, error
	resData, resHeader, httpStatus, resErr := c.fetchJSON(ctx, method, endpoint, reqPayload)
	if resErr == nil {
		return resData, resHeader, httpStatus, nil
	}

	shouldRetry := false
	if resErr.Code != 0 {
		shouldRetry = resErr.IsSameCode(errs.AccessTokenNotFound)
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
	return c.fetchJSON(ctx, method, endpoint, reqPayload)
}

func (c *ExternalFWAPIClient) fetchPDFBytes(ctx context.Context, method string, endpoint string, reqPayload *RequestPayload) ([]byte, http.Header, int, *errs.Error) { // pdf bytes, response header, http status, error
	// Build the actual payload: caller's body + caller's headers + framework's Accept (forced).
	// Accept is fetchPDFBytes's contract — caller can't override it.
	actualReqPayload := &RequestPayload{Headers: http.Header{}}
	if reqPayload != nil {
		for k, vs := range reqPayload.Headers {
			actualReqPayload.Headers[k] = vs
		}
		actualReqPayload.BodyProvider = reqPayload.BodyProvider
	}
	actualReqPayload.Headers.Set("Accept", "application/pdf")

	upstrRes, status, resErr := c.doRequest(ctx, method, endpoint, actualReqPayload)
	if resErr != nil {
		return nil, nil, status, resErr
	}
	defer func() { _ = upstrRes.Body.Close() }()

	pdfData, err := io.ReadAll(upstrRes.Body)
	if err != nil {
		return nil, nil, http.StatusBadGateway, errs.Upstream.Wrap(err)
	}
	return pdfData, upstrRes.Header, http.StatusOK, nil
}

func (c *ExternalFWAPIClient) FetchPDFBytesRetriable(ctx context.Context, method string, endpoint string, reqPayload *RequestPayload) ([]byte, http.Header, int, *errs.Error) { // pdf bytes, response header, http status, error
	pdfData, resHeader, httpStatus, resErr := c.fetchPDFBytes(ctx, method, endpoint, reqPayload)
	if resErr == nil {
		return pdfData, resHeader, httpStatus, nil
	}

	shouldRetry := false
	if resErr.Code != 0 {
		shouldRetry = resErr.IsSameCode(errs.AccessTokenNotFound)
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
	return c.fetchPDFBytes(ctx, method, endpoint, reqPayload)
}

func (c *ExternalFWAPIClient) fetchPDFStream(ctx context.Context, method string, endpoint string, reqPayload *RequestPayload) (io.ReadCloser, http.Header, int, *errs.Error) { // stream, response header, http status, error
	// Build the actual payload: caller's body + caller's headers + framework's Accept (forced).
	// Accept is fetchPDFStream's contract — caller can't override it.
	actualReqPayload := &RequestPayload{Headers: http.Header{}}
	if reqPayload != nil {
		for k, vs := range reqPayload.Headers {
			actualReqPayload.Headers[k] = vs
		}
		actualReqPayload.BodyProvider = reqPayload.BodyProvider
	}
	actualReqPayload.Headers.Set("Accept", "application/pdf")

	upstrRes, status, resErr := c.doRequest(ctx, method, endpoint, actualReqPayload)
	if resErr != nil {
		return nil, nil, status, resErr
	}
	// Success: Response Body NOT closed here — caller owns the stream and must close it.
	return upstrRes.Body, upstrRes.Header, http.StatusOK, nil
}

func (c *ExternalFWAPIClient) FetchPDFStreamRetriable(ctx context.Context, method string, endpoint string, reqPayload *RequestPayload) (io.ReadCloser, http.Header, int, *errs.Error) { // stream, response header, http status, error
	stream, resHeader, httpStatus, resErr := c.fetchPDFStream(ctx, method, endpoint, reqPayload)
	if resErr == nil {
		return stream, resHeader, httpStatus, nil
	}

	shouldRetry := false
	if resErr.Code != 0 {
		shouldRetry = resErr.IsSameCode(errs.AccessTokenNotFound)
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
	return c.fetchPDFStream(ctx, method, endpoint, reqPayload)
}
