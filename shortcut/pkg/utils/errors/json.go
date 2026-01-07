package errorsutils

type JSONError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

func NewJSONError(message string) JSONError {
	return JSONError{
		Message: message,
	}
}

func NewJSONErrorWithCode(code, message string) JSONError {
	return JSONError{
		Code:    code,
		Message: message,
	}
}
