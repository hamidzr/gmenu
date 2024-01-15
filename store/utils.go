package store

import (
	"os"
	"path/filepath"

	"github.com/hamidzr/gmenu/constants"
)

func ConfigDir() string {
	return filepath.Join(os.Getenv("HOME"), ".config", constants.ProjectName)
}

func CacheDir() string {
	return filepath.Join(os.Getenv("HOME"), ".cache", constants.ProjectName)
}
