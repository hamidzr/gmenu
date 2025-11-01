package main

import (
	"os"

	"github.com/hamidzr/gmenu/internal/cli"
	"github.com/hamidzr/gmenu/internal/logger"
	"github.com/hamidzr/gmenu/model"
	"github.com/sirupsen/logrus"
)

func run() model.ExitCode {
	// Suppress noisy macOS subsystem logs without discarding our own output.
	_ = os.Setenv("OS_ACTIVITY_MODE", "disable")

	logger.SetupLogger()
	cmd := cli.InitCLI()
	if err := cmd.Execute(); err != nil {
		logrus.WithError(err).Error("gmenu exited with error")
		return model.UnknownError
	}
	return model.NoError
}

func main() {
	cleanup := startProfiling()
	exitCode := run()
	cleanup()
	os.Exit(int(exitCode))
}
