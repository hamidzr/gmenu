package store

import (
	"os"
	"path/filepath"

	"github.com/hamidzr/gmenu/constants"
)

func ConfigDir(menuID string) string {
	return filepath.Join(os.Getenv("HOME"), ".config", constants.ProjectName, menuID)
}

func CacheDir(menuID string) string {
	return filepath.Join(os.Getenv("HOME"), ".cache", constants.ProjectName, menuID)
}
