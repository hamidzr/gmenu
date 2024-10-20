package core

import (
	"fmt"
	"os"
)

const (
	unsetInt = -1
)

func createPidFile(name string) (string, error) {
	dir := os.TempDir()
	pidFile := fmt.Sprintf("%s/%s.pid", dir, name)
	if _, err := os.Stat(pidFile); err == nil {
		fmt.Println("Another instance of gmenu is already running. Exiting.")
		fmt.Println("If this is not the case, please delete the pid file:", pidFile)
		return "", fmt.Errorf("pid file already exists")

	}
	f, err := os.Create(pidFile)
	if err != nil {
		fmt.Println("Failed to create pid file")
		if ferr := f.Close(); ferr != nil {
			fmt.Println("Failed to close pid file:", ferr)
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
