package common

import (
	"fmt"
)

// Profile is the struct that will store the profile of the person who logged in
type Profile struct {
	ID, DisplayName, ImageURL string
}

// AppError is the error struct
// http://blog.golang.org/error-handling-and-go
type AppError struct {
	Error   error
	Message string
	Code    int
}

// AppErrorf creates and returns an AppError object
func AppErrorf(err error, format string, v ...interface{}) *AppError {
	return &AppError{
		Error:   err,
		Message: fmt.Sprintf(format, v...),
		Code:    500,
	}
}
