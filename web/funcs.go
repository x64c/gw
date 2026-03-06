package web

import (
	"net/http"
)

// ShallowCloneClient returns a shallow copy of the given *http.Client.
// The clone shares the Transport and other pointer fields, but config fields like Timeout are independent.
// Syntax update in Go 1.26:
// replace pointer to local variable 'c' with new()
// := new(*base)
func ShallowCloneClient(base *http.Client) *http.Client {
	c := *base
	return &c
}
