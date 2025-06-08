package core

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	defaultPidFileNem = "gmenu"
)

func removePidFile(name string) error {
	if name == "" {
		name = defaultPidFileNem
	}
	dir := os.TempDir()
	pidFile := fmt.Sprintf("%s/%s.pid", dir, name)
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return fmt.Errorf("pid file does not exist")
	}
	err := os.Remove(pidFile)
	if err != nil {
		logrus.Error("Failed to remove pid file:", err)
		return err
	}
	logrus.Info("Pid file removed successfully:", pidFile)
	return nil
}

func createPidFile(name string) (string, error) {
	if name == "" {
		name = defaultPidFileNem
	}
	dir := os.TempDir()
	pidFile := fmt.Sprintf("%s%s.pid", dir, name)
	if _, err := os.Stat(pidFile); err == nil {
		logrus.Warn("Another instance of gmenu is already running. Exiting.")
		logrus.Warn("If this is not the case, please delete the pid file:", pidFile)
		return "", fmt.Errorf("pid file already exists")

	}
	f, err := os.Create(pidFile)
	if err != nil {
		logrus.Error("Failed to create pid file")
		if ferr := f.Close(); ferr != nil {
			logrus.Error("Failed to close pid file:", ferr)
		}
		return "", err
	}
	return pidFile, f.Close()
}

// canBeHighlighted returns true if the menu item can be highlighted
// programmatically via exiting fayne interface.
func canBeHighlighted(entry string) bool {
	// TODO: find a better way to select all on searchEnty.
	for _, c := range entry {
		if !(c >= 'a' && c <= 'z' ||
			c >= 'A' && c <= 'Z' ||
			c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}
