package cantabular

import (
	"errors"
	"net/http"
)

// Error is the package's error type
type Error struct{
	err error
	statusCode int
}

// Error implements the standard Go error
func (e *Error) Error() string {
	return e.err.Error()
}

// Unwrap implements Go error unwrapping
func (e *Error) Unwrap() error{
	return e.err
}

// Code returns the statusCode returned by Cantabular.
// Begrudingly called Code rather than StatusCode but this is
// how it is named elsewhere accross ONS services and is more useful
// being consistent
func (e *Error) Code() int{
	return e.statusCode
}

// StatusCode is a callback function that allows you to extract
// a status code from an error, or returns 500 as a default
func StatusCode(err error) int{
	var cerr coder
	if errors.As(err, &cerr){
		return cerr.Code()
	}

	return http.StatusInternalServerError
}