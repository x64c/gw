package responses

type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`  // application-level logic code
	Cause   error  `json:"-"`    // original error, not serialized
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Cause
}
