package model

import "fmt"

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
