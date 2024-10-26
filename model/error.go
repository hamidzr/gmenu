package model

import (
	"errors"
	"fmt"
)

type ExitCode int

func (e ExitCode) String() string {
	return fmt.Sprintf("Exit code %d", e)
}

func (e ExitCode) Error() string {
	return e.String()
}

const (
	Unset ExitCode = -1
)
const (
	NoError ExitCode = iota
	UnknownError
	UserCanceled
)

var (
	// CustomUserEntry is when the user inputs and pushes an entry through that doesn't exist
	// and gmenu is not set to accept it.
	CustomUserEntry = errors.New("unmatched user entry")
)
