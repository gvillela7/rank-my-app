package handlers

import (
	"errors"
	"fmt"
	"net/http"
)

type HTTPError struct {
	Code    int
	Message string
	Err     error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func NewHTTPError(code int, message string, err error) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func BadRequestError(message string, err error) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, message, err)
}

func InternalServerError(message string, err error) *HTTPError {
	return NewHTTPError(http.StatusInternalServerError, message, err)
}

func NotFoundError(message string) *HTTPError {
	return NewHTTPError(http.StatusNotFound, message, nil)
}

func UnauthorizedError(message string) *HTTPError {
	return NewHTTPError(http.StatusUnauthorized, message, nil)
}

func GetHTTPError(err error) (*HTTPError, bool) {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr, true
	}
	return nil, false
}
