package store

import (
	"os"
	"path/filepath"
)

func ConfigDir(namespace string) string {
	return filepath.Join(os.Getenv("HOME"), ".config", namespace)
}

func CacheDir(namespace string) string {
	return filepath.Join(os.Getenv("HOME"), ".cache", namespace)
}
