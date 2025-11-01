package model

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExitErrorWrapping(t *testing.T) {
	rootErr := errors.New("boom")
	exitErr := NewExitError(UnknownError, rootErr)

	require.NotNil(t, exitErr)
	assert.Equal(t, rootErr, exitErr.Err)
	assert.Contains(t, exitErr.Error(), "Exit code 1")

	code, cause := ExitCodeFromError(exitErr)
	assert.Equal(t, UnknownError, code)
	assert.Equal(t, rootErr, cause)
}

func TestExitCodeFromNonExitError(t *testing.T) {
	plainErr := errors.New("plain")

	code, cause := ExitCodeFromError(plainErr)
	assert.Equal(t, UnknownError, code)
	assert.Equal(t, plainErr, cause)
}

func TestExitCodeFromNilExitError(t *testing.T) {
	exitErr := NewExitError(UserCanceled, nil)

	code, cause := ExitCodeFromError(exitErr)
	assert.Equal(t, UserCanceled, code)
	assert.Nil(t, cause)
}
