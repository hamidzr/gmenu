package main

import (
	"testing"

	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/assert"
)

func TestRunRequiresArgument(t *testing.T) {
	err := run([]string{"gmenu-visual-test"})
	assert.Error(t, err)

	code, cause := model.ExitCodeFromError(err)
	assert.Equal(t, model.UnknownError, code)
	assert.Nil(t, cause)
}

func TestRunUnknownTestType(t *testing.T) {
	err := run([]string{"gmenu-visual-test", "not-real"})
	assert.Error(t, err)

	code, cause := model.ExitCodeFromError(err)
	assert.Equal(t, model.UnknownError, code)
	assert.Nil(t, cause)
}
