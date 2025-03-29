package main

import (
	"os"

	"github.com/hamidzr/gmenu/internal/cli"
	"github.com/hamidzr/gmenu/internal/logger"
	"github.com/sirupsen/logrus"
)

func main() {
	// Suppress system debug output
	os.Setenv("OS_ACTIVITY_MODE", "debug")
	os.Stderr.Close()
	os.Stderr, _ = os.Open(os.DevNull)

	cmd := cli.InitCLI()
	logger.SetupLogger()
	if err := cmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
