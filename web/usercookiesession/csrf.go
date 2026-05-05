package usercookiesession

import (
	"context"
	"fmt"
	"html/template"
)

const (
	CSRFHeaderName    = "X-CSRF-Token"
	CSRFFormFieldName = "_csrf_token"
)

// CSRFTokenFromContext returns the CSRF token attached to the request's cookie
// session, or ("", false) if no SessionData is present.
func CSRFTokenFromContext[UID comparable](ctx context.Context) (string, bool) {
	sd, ok := FromContext[UID](ctx)
	if !ok {
		return "", false
	}
	return sd.CSRFTkn, true
}

// CSRFHiddenInputHTML returns the hidden <input> element carrying the CSRF
// token under CSRFFormFieldName, suitable for embedding in <form> templates.
//
// Returns template.HTML so html/template renders it raw (without escaping the
// angle brackets).
//
// Usage
// - Inject under the conventional key ".csrf":
//
//	tkn, _ := usercookiesession.CSRFTokenFromContext[int64](r.Context())
//	data := map[string]any{
//	    "csrf": usercookiesession.CSRFHiddenInputHTML(tkn),
//	    // ... other template data ...
//	}
//	tpl1.Execute(w, data)
//
// ----
// - Embed it inside the <form>:
//
//	<form ...>
//	    {{ .csrf }}
//	    <!-- other fields -->
//	    <button type="submit">Submit</button>
//	</form>
//
// For AJAX / JS / JSON contexts (X-CSRF-Token header, etc.), use
// CSRFTokenFromContext to get the bare token string instead.
func CSRFHiddenInputHTML(token string) template.HTML {
	return template.HTML(fmt.Sprintf(`<input type="hidden" name=%q value=%q>`, CSRFFormFieldName, token))
}
