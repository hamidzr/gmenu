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

// TODO cache and cofnig logic are the same.

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
	logger.Info("loaded cache", "cache", cache)
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
