package web

// ExternalFWAPIConf for a client of an External API built on this framework (Framework API)
type ExternalFWAPIConf struct {
	Host                       string            `json:"host"`
	ClientID                   string            `json:"client_id"`                 // ID of this App as a ExternalAPIClient of the MainBackendAPI
	ReissueAccessTokenEndpoint string            `json:"reissue_access_token"`      // path after host
	ReissueIdTokenEndpoint     string            `json:"reissue_id_token"`          // path after host
	JwksURL                    string            `json:"jwks_url"`                  // full url
	VerifyAuthCodeEndpoints    map[string]string `json:"verify_external_auth_code"` // path after host
}
