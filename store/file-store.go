package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// TODO: we shoudn't need to differentiate between these two or special case them
var logger = logrus.New()

// FileStore is a store that saves data to files in the cache and config directories.
type FileStore[Cache any, Cfg any] struct {
	cacheDir  string
	configDir string
	format    string
}

func (fs FileStore[Cache, Cfg]) Marshal(data any) ([]byte, error) {
	if fs.format == "json" {
		return json.Marshal(data)
	}
	return yaml.Marshal(data)
}

func (fs FileStore[Cache, Cfg]) Unmarshal(data []byte, v any) error {
	if fs.format == "json" {
		return json.Unmarshal(data, v)
	}
	return yaml.Unmarshal(data, v)
}

// buildFilePath creates a file path with the given directory and filename
func (fs FileStore[C, Cfg]) buildFilePath(dir, name string) string {
	return filepath.Join(dir, name+"."+fs.format)
}

// saveData is a generic helper for saving data to a file
func (fs FileStore[C, Cfg]) saveData(data any, filePath string) error {
	serialized, err := fs.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, serialized, 0o644)
}

// loadData is a generic helper for loading data from a file
func (fs FileStore[C, Cfg]) loadData(filePath string, target any, allowMissing bool) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if allowMissing {
			return nil // Return zero value for cache
		}
		return err // Return error for config
	}
	serialized, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return fs.Unmarshal(serialized, target)
}

// NewFileStore initializes a new FileStore with directories for cache and config.
func NewFileStore[Cache any, Cfg any](namespace []string, format string) (*FileStore[Cache, Cfg], error) {
	cacheDir := CacheDir("")
	configDir := ConfigDir("")
	if format != "json" && format != "yaml" {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
	for _, dir := range namespace {
		cacheDir = filepath.Join(cacheDir, dir)
		configDir = filepath.Join(configDir, dir)
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return nil, err
	}
	return &FileStore[Cache, Cfg]{
		cacheDir:  cacheDir,
		configDir: configDir,
		format:    format,
	}, nil
}

func (fs FileStore[C, Cfg]) Load() (C, Cfg, error) {
	cache, cacheErr := fs.LoadCache()
	logger.Info("loaded cache file at ", fs.cacheFilePath())
	config, configErr := fs.LoadConfig()
	var err error
	if cacheErr != nil {
		err = cacheErr
	}
	if configErr != nil {
		err = configErr
	}
	return cache, config, err
}

func (fs FileStore[C, Cfg]) Save(data C, config Cfg) error {
	if err := fs.SaveCache(data); err != nil {
		return err
	}
	return fs.SaveConfig(config)
}
