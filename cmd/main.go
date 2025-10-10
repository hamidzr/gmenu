package main

import (
	"os"

	"github.com/hamidzr/gmenu/internal/cli"
	"github.com/hamidzr/gmenu/internal/logger"
	"github.com/sirupsen/logrus"
)

func main() {
	// Suppress noisy macOS subsystem logs without discarding our own output.
	_ = os.Setenv("OS_ACTIVITY_MODE", "disable")

	logger.SetupLogger()
	cmd := cli.InitCLI()
	if err := cmd.Execute(); err != nil {
		logrus.WithError(err).Error("gmenu exited with error")
		os.Exit(1)
	}
}
