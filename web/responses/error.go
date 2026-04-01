package responses

type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`  // logic error code. See: codes.go
	Cause   error  `json:"-"`    // original error, not serialized
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// WithDetail returns a new Error with the detail appended to the message.
func (e *Error) WithDetail(detail string) *Error {
	return &Error{Code: e.Code, Message: e.Message + ": " + detail, Cause: e.Cause}
}
