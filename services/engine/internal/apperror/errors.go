package apperror

import "net/http"

type AppError struct {
	Code    int    // HTTP status code
	Message string // user-facing message
	Err     error  // internal cause (never sent to client)
}

func (e *AppError) Error() string {
	return e.Message
}

// Constructors

func Internal(err error) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
		Err:     err,
	}
}
