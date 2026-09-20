package cli

import (
	"errors"
	"fmt"
)

// CLIError is a typed error with an exit code.
type CLIError struct {
	Code int
	Msg  string
}

func (e *CLIError) Error() string {
	return e.Msg
}

// Convenience constructors for common exit codes.
func userErr(format string, args ...interface{}) error {
	return &CLIError{Code: 2, Msg: fmt.Sprintf(format, args...)}
}

func netErr(format string, args ...interface{}) error {
	return &CLIError{Code: 3, Msg: fmt.Sprintf(format, args...)}
}

// IsCLIError checks if err is a *CLIError.
func IsCLIError(err error) (*CLIError, bool) {
	var ce *CLIError
	if errors.As(err, &ce) {
		return ce, true
	}
	return nil, false
}
