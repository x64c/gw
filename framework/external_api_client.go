package framework

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
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
func (c *ExternalAPIClient) RequestJSON(ctx context.Context, accessToken string, method string, endpoint string) (*http.Response, error) {
	upstrUrl := c.Conf.Host + endpoint
	upstrReq, err := http.NewRequestWithContext(ctx, method, upstrUrl, nil)
	if err != nil {
		return nil, err
	}

	upstrReq.Header.Set("Client-Id", c.Conf.ClientID)
	upstrReq.Header.Set("Authorization", "Bearer "+accessToken)
	upstrReq.Header.Set("Content-Type", "application/json")
	upstrReq.Header.Set("Accept", "application/json")

	upstrRes, err := c.Do(upstrReq)
	if err != nil {
		return nil, err
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

func (c *ExternalAPIClient) UpdateTokens(ctx context.Context) (*security.AccessTokenPair, int, error) {
	log.Printf("[DEBUG] preparing to request to update tokens for %q", c.ApiID)
	sessionID, ok := usercookiesession.SessionIDFromContext(ctx)
	if !ok {
		return nil, http.StatusUnauthorized, errors.New("no session iD in the context")
	}
	log.Printf("[DEBUG] sessionID %q for UpdateTokens", sessionID)
	cookieSessionMgr := c.Core.UserCookieSessionManager

	uidStr, err := cookieSessionMgr.SessionIDToUIDStrFromKVDB(ctx, sessionID)
	if err != nil {
		return nil, http.StatusUnauthorized, errors.New("no session info for the session id")
	}
	refreshToken, err := cookieSessionMgr.FetchExternalRefreshToken(ctx, sessionID, c.ApiID)

	log.Printf("[DEBUG] requesting to update tokens with refresh token %q", refreshToken)

	upstrRes, err := c.RequestReissueAccessTokenWithRefreshTokenAndUserID(ctx, refreshToken, uidStr)
	if err != nil {
		return nil, http.StatusServiceUnavailable, err
	}
	defer func() { _ = upstrRes.Body.Close() }()

	log.Printf("[DEBUG] got UpdateTokens response from upstream with status code %d", upstrRes.StatusCode)

	if upstrRes.StatusCode != http.StatusOK {
		return nil, upstrRes.StatusCode, errors.New(upstrRes.Status)
	}

	newTokenPair := security.AccessTokenPair{}
	if err = json.UnmarshalRead(upstrRes.Body, &newTokenPair); err != nil {
		log.Print("[DEBUG] failed to unmarshal new token pair")
		return nil, upstrRes.StatusCode, err
	}

	// Update Tokens
	if err = cookieSessionMgr.StoreExternalTokenPairInKVDB(ctx, sessionID, c.ApiID, newTokenPair.AccessToken, newTokenPair.RefreshToken); err != nil {
		log.Printf("[DEBUG] failed to store new token pair in kvdb: %q, %q", newTokenPair.AccessToken, newTokenPair.RefreshToken)
		return nil, http.StatusInternalServerError, err
	}
	log.Printf("[DEBUG] tokens updated: %q, %q", newTokenPair.AccessToken, newTokenPair.RefreshToken)

	return &newTokenPair, http.StatusOK, nil
}

func (c *ExternalAPIClient) fetchJSON(ctx context.Context, method string, endpoint string) (any, int, error) { // data, http.StatusCode, error
	sessionID, ok := usercookiesession.SessionIDFromContext(ctx)
	if !ok {
		return nil, http.StatusBadRequest, errors.New("no session iD in the context")
	}
	accessToken, err := c.Core.UserCookieSessionManager.FetchExternalAccessToken(ctx, sessionID, c.ApiID)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("no access token for the api %q in the session", c.ApiID)
	}
	upstrRes, err := c.RequestJSON(ctx, accessToken, method, endpoint)
	if err != nil {
		return nil, http.StatusServiceUnavailable, err
	}
	defer func() { _ = upstrRes.Body.Close() }()

	if upstrRes.StatusCode == http.StatusNotFound {
		// 404 not found -> raw error message sent before wrapped into JSON
		// ToDo: Handle the case: found the endpoint, but requested data resource not found
		return nil, http.StatusNotFound, responses.HTTPErrorNotFound
	}

	if upstrRes.StatusCode != http.StatusOK {
		var resMsg responses.Message
		err = json.UnmarshalRead(upstrRes.Body, &resMsg)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.New("failed to unmarshal server message")
		}
		return resMsg, upstrRes.StatusCode, nil
	}

	// Now we expect JSON data response body
	var resData any
	if err = json.UnmarshalRead(upstrRes.Body, &resData); err != nil {
		return nil, upstrRes.StatusCode, err
	}

	return resData, upstrRes.StatusCode, nil
}

func (c *ExternalAPIClient) FetchJSONRetriable(ctx context.Context, method string, endpoint string) (any, int, error) { // data, http.StatusCode, error
	log.Print("[DEBUG] preparing to request json to main backend")
	resData, httpStatusCode, err := c.fetchJSON(ctx, method, endpoint)
	if err != nil {
		// internal error
		log.Printf("[DEBUG] [ERROR] http code: %d _ internal error %v", httpStatusCode, err)
		return resData, httpStatusCode, err
	}
	if httpStatusCode == http.StatusUnauthorized {
		log.Printf("[DEBUG] upstream response http code: %d", httpStatusCode)
		// resData is expected to responses.Message
		resMessage, ok := resData.(responses.Message)
		if !ok {
			// no idea about the data. just return the data
			return resData, httpStatusCode, nil
		}
		if resMessage.Code != reason.AccessTokenExpired {
			// got error message but not access token expired. return the message
			log.Printf("[DEBUG] upstream response logic error code: %d", resMessage.Code)
			return resMessage, httpStatusCode, nil
		}
		log.Printf("[DEBUG] upstream response logic error: `AccessTokenExpired` (%d)", resMessage.Code)
		// got "access token expired" message, request to update tokens
		_, httpStatusCode, err = c.UpdateTokens(ctx)
		if err != nil || httpStatusCode != http.StatusOK {
			log.Printf("[DEBUG] [http %d] failed to update tokens %v", httpStatusCode, err)
			return nil, httpStatusCode, err
		}
		// retry with new access token
		log.Print("[DEBUG] retrying to fetch json")
		resData, httpStatusCode, err = c.fetchJSON(ctx, method, endpoint)
	}
	log.Print("[DEBUG] got json response")
	return resData, httpStatusCode, err
}

func (c *ExternalAPIClient) fetchPDFBytes(ctx context.Context, method string, endpoint string) (
	any, int, http.Header, error,
) { // bytes or errMsg, http.StatusCode, http.Header, error
	sessionID, ok := usercookiesession.SessionIDFromContext(ctx)
	if !ok {
		return nil, http.StatusBadRequest, nil, errors.New("no session iD in the context")
	}
	accessToken, err := c.Core.UserCookieSessionManager.FetchExternalAccessToken(ctx, sessionID, c.ApiID)
	if err != nil {
		return nil, http.StatusBadRequest, nil, fmt.Errorf("no access token for the api %q in the session", c.ApiID)
	}
	upstrUrl := c.Conf.Host + endpoint
	upstrReq, err := http.NewRequestWithContext(ctx, method, upstrUrl, nil)
	if err != nil {
		return nil, http.StatusBadRequest, nil, err
	}

	upstrReq.Header.Set("Client-Id", c.Conf.ClientID)
	upstrReq.Header.Set("Authorization", "Bearer "+accessToken)
	upstrReq.Header.Set("Content-Type", "application/json")
	upstrReq.Header.Set("Accept", "application/pdf")

	upstrRes, err := c.Do(upstrReq)
	if err != nil {
		return nil, http.StatusServiceUnavailable, nil, err
	}
	defer func() { _ = upstrRes.Body.Close() }()

	if upstrRes.StatusCode == http.StatusNotFound {
		// 404 not found -> raw error message sent before wrapped into JSON
		return nil, http.StatusNotFound, upstrRes.Header, responses.HTTPErrorNotFound
	}

	if upstrRes.StatusCode != http.StatusOK {
		var resMsg responses.Message
		err = json.UnmarshalRead(upstrRes.Body, &resMsg)
		if err != nil {
			return nil, http.StatusInternalServerError, upstrRes.Header, errors.New("failed to unmarshal server message")
		}
		return resMsg, upstrRes.StatusCode, upstrRes.Header, nil
	}

	// Now we expect PDF response body
	pdfData, err := io.ReadAll(upstrRes.Body)
	if err != nil {
		return nil, upstrRes.StatusCode, upstrRes.Header, err
	}

	return pdfData, upstrRes.StatusCode, upstrRes.Header, nil
}

func (c *ExternalAPIClient) FetchPDFBytesRetriable(ctx context.Context, method string, endpoint string) (
	any, int, http.Header, error,
) { // bytes or errMsg, http.StatusCode, http.Header, error
	pdfData, httpStatusCode, resHeader, err := c.fetchPDFBytes(ctx, method, endpoint)
	if err != nil {
		// internal error
		return pdfData, httpStatusCode, resHeader, err
	}
	if httpStatusCode == http.StatusUnauthorized {
		// pdfData is expected to responses.Message
		resMessage, ok := pdfData.(responses.Message)
		if !ok {
			// no idea about the data. just return the data
			return pdfData, httpStatusCode, resHeader, nil
		}
		if resMessage.Code != reason.AccessTokenExpired {
			// got error message but not access token expired. return the message
			return resMessage, httpStatusCode, resHeader, nil
		}
		// got "access token expired" message, request to update tokens
		_, httpStatusCode, err = c.UpdateTokens(ctx)
		if err != nil || httpStatusCode != http.StatusOK {
			return nil, httpStatusCode, resHeader, err
		}
		// retry with new access token
		pdfData, httpStatusCode, resHeader, err = c.fetchPDFBytes(ctx, method, endpoint)
	}
	return pdfData, httpStatusCode, resHeader, err
}

func (c *ExternalAPIClient) fetchPDFStream(ctx context.Context, method string, endpoint string) (
	io.ReadCloser, *responses.Message, int, http.Header, error,
) { // stream, json msg, http.StatusCode, http.Header, error
	sessionID, ok := usercookiesession.SessionIDFromContext(ctx)
	if !ok {
		return nil, nil, http.StatusBadRequest, nil, errors.New("no session iD in the context")
	}
	accessToken, err := c.Core.UserCookieSessionManager.FetchExternalAccessToken(ctx, sessionID, c.ApiID)
	if err != nil {
		return nil, nil, http.StatusBadRequest, nil, fmt.Errorf("no access token for the api %q in the session", c.ApiID)
	}
	upstrUrl := c.Conf.Host + endpoint
	upstrReq, err := http.NewRequestWithContext(ctx, method, upstrUrl, nil)
	if err != nil {
		return nil, nil, http.StatusBadRequest, nil, err
	}

	upstrReq.Header.Set("Client-Id", c.Conf.ClientID)
	upstrReq.Header.Set("Authorization", "Bearer "+accessToken)
	upstrReq.Header.Set("Content-Type", "application/json")
	upstrReq.Header.Set("Accept", "application/pdf")

	upstrRes, err := c.Do(upstrReq)
	if err != nil {
		return nil, nil, http.StatusServiceUnavailable, nil, err
	}

	if upstrRes.StatusCode == http.StatusOK {
		// SUCCESS: stream PDF
		return upstrRes.Body, nil, upstrRes.StatusCode, upstrRes.Header, nil
	}

	// ERROR PATH — must consume & close body
	defer func() { _ = upstrRes.Body.Close() }()

	if upstrRes.StatusCode == http.StatusNotFound {
		// 404 not found -> raw error message sent before wrapped into JSON
		return nil, nil, http.StatusNotFound, upstrRes.Header, responses.HTTPErrorNotFound
	}

	var resMsg responses.Message
	if err := json.UnmarshalRead(upstrRes.Body, &resMsg); err != nil {
		return nil, nil, http.StatusInternalServerError, upstrRes.Header, err
	}
	return nil, &resMsg, upstrRes.StatusCode, upstrRes.Header, nil
}

// FetchPDFStreamRetriable is a wrapper to retry when the access token is expired
// 1. Try fetchPDFStream
// 2. If success, return the stream immediately
// 3. If JSON error is returned, inspect it
// 4. If error code is AccessTokenExpired, refresh token and retry once
func (c *ExternalAPIClient) FetchPDFStreamRetriable(ctx context.Context, method string, endpoint string) (
	io.ReadCloser, *responses.Message, int, http.Header, error,
) { // stream, json msg, http.StatusCode, http.Header, error
	// ---- first attempt ----
	stream, resMsg, status, hdr, err := c.fetchPDFStream(ctx, method, endpoint)
	if err != nil {
		// transport / internal error
		return nil, nil, status, hdr, err
	}
	if stream != nil {
		// success: stream PDF
		return stream, nil, status, hdr, nil
	}
	if resMsg == nil {
		// non-JSON error (e.g. 404 already handled upstream)
		return nil, nil, status, hdr, nil
	}
	// ---- inspect JSON error ----
	if status != http.StatusUnauthorized || resMsg.Code != reason.AccessTokenExpired {
		// not a retryable auth error
		return nil, resMsg, status, hdr, nil
	}
	// ---- refresh access token ----
	_, status, err = c.UpdateTokens(ctx)
	if err != nil || status != http.StatusOK {
		return nil, nil, status, hdr, err
	}
	// ---- retry once with fresh token ----
	stream, resMsg, status, hdr, err = c.fetchPDFStream(ctx, method, endpoint)
	if err != nil {
		return nil, nil, status, hdr, err
	}
	return stream, resMsg, status, hdr, nil
}
