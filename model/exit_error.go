package model

import (
	"errors"
	"fmt"
)

// ExitError represents an application exit that should propagate to the process
// exit code without abruptly terminating the program. It enables lower layers
// to communicate the desired exit code to callers while still allowing cleanup
// logic to run.
type ExitError struct {
	Code ExitCode
	Err  error
}

// Error implements the error interface.
func (e *ExitError) Error() string {
	if e == nil {
		return ""
	}

	if e.Err == nil {
		return e.Code.String()
	}

	return fmt.Sprintf("%s: %v", e.Code.String(), e.Err)
}

// Unwrap exposes the underlying error.
func (e *ExitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// NewExitError constructs an ExitError with the provided code and cause.
func NewExitError(code ExitCode, err error) *ExitError {
	return &ExitError{Code: code, Err: err}
}

// ExitCodeFromError extracts an ExitCode from the provided error. If the error
// is an ExitError, its embedded code is returned. Otherwise UnknownError is
// returned together with the original error for logging.
func ExitCodeFromError(err error) (ExitCode, error) {
	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		return exitErr.Code, exitErr.Err
	}
	return UnknownError, err
}
