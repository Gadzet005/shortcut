package shortcut

import "net/http"

var (
	ErrItemNotFound = NewError(http.StatusBadRequest, "required item not found")
)

func NewError(code int, message string) *HandlerError {
	return &HandlerError{
		StatusCode: code,
		Message:    message,
	}
}

func NewErrorWithCause(code int, message string, cause error) *HandlerError {
	return &HandlerError{
		StatusCode: code,
		Message:    message,
		Err:        cause,
	}
}

type HandlerError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *HandlerError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *HandlerError) Unwrap() error {
	return e.Err
}
