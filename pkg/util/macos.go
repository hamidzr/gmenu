package util

import (
	"context"
	"os/exec"

	"github.com/pkg/errors"
)

func MinimizeWindow(ctx context.Context, windowTitle string) error {
	appname := "TODO"
	script := `tell application "System Events" to set miniaturized of window "` + windowTitle + `" of application "` + +`" to true`

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	err := cmd.Run()
	if err != nil {
		// read the stderr
		if exitErr, ok := err.(*exec.ExitError); ok {
			return errors.Wrapf(err, "osascript stderr: %s", string(exitErr.Stderr))
		}
		return errors.Wrap(err, "failed to run osascript")
	}
	return nil
}
