package web

import (
	"io"
	"net/http"
)

// RequestPayload bundles caller-provided inputs that ExternalFWAPIClient methods
// fold into the outgoing Request: extra headers and the Request Body provider.
// Pass nil to send a body-less Request with no extra headers.
type RequestPayload struct {
	Headers      http.Header               // additional headers merged into the Request (Content-Type, custom headers, etc.). Framework auth headers (Authorization, Client-Id) are set last and always win.
	BodyProvider func() (io.Reader, error) // closure to provide a fresh Request Body reader on each call. Called once per attempt so framework retries (token refresh) get a fresh reader. Also wired to stdlib's req.GetBody so HTTP redirects / HTTP/2 retries replay correctly. Pass nil for a body-less Request.
}
