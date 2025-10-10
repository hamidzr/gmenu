package core

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	defaultPidFileName = "gmenu"
)

func RemovePidFile(name string) error {
	if name == "" {
		name = defaultPidFileName
	}
	dir := os.TempDir()
	pidFile := fmt.Sprintf("%s/%s.pid", dir, name)
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		logrus.Debug("Pid file already absent, skipping remove:", pidFile)
		return nil
	}
	err := os.Remove(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debug("Pid file disappeared before remove:", pidFile)
			return nil
		}
		logrus.Error("Failed to remove pid file:", err)
		return err
	}
	logrus.Info("Pid file removed successfully:", pidFile)
	return nil
}

func createPidFile(name string) (string, error) {
	if name == "" {
		name = defaultPidFileName
	}
	dir := os.TempDir()
	pidFile := fmt.Sprintf("%s/%s.pid", dir, name)
	if _, err := os.Stat(pidFile); err == nil {
		logrus.Warn("Another instance of gmenu is already running. Exiting.")
		logrus.Warn("If this is not the case, please delete the pid file:", pidFile)
		return "", fmt.Errorf("gmenu-%s pid file already exists", name)
	}
	f, err := os.Create(pidFile)
	if err != nil {
		logrus.Error("Failed to create gmenu pid file")
		return "", err
	}
	// ensure file is closed before returning
	if closeErr := f.Close(); closeErr != nil {
		logrus.Error("Failed to close pid file:", closeErr)
		// attempt cleanup on close failure
		if removeErr := os.Remove(pidFile); removeErr != nil {
			logrus.Error("Failed to clean up pid file after close error:", removeErr)
		}
		return "", closeErr
	}
	return pidFile, nil
}

// canBeHighlighted returns true if the menu item can be highlighted
// programmatically via existing fyne interface.
// Currently restricted to alphanumeric characters due to fyne limitations.
func canBeHighlighted(entry string) bool {
	if entry == "" {
		return true
	}
	for _, c := range entry {
		if !isAlphaNumeric(c) {
			return false
		}
	}
	return true
}

// isAlphaNumeric returns true if the character is alphanumeric
func isAlphaNumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9')
}
