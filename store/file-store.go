package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FileStore is a store that saves data to files in the cache and config directories.
type FileStore[C any, Cfg any] struct {
	cacheDir  string
	configDir string
}

// NewFileStore initializes a new FileStore with directories for cache and config.
func NewFileStore[C any, Cfg any](namespace []string) (*FileStore[C, Cfg], error) {
	cacheDir := CacheDir("")
	configDir := ConfigDir("")
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
	return &FileStore[C, Cfg]{
		cacheDir:  cacheDir,
		configDir: configDir,
	}, nil
}

// SaveCache serializes and saves the cache data to a file.
func (fs *FileStore[C, Cfg]) SaveCache(data C) error {
	serialized, err := json.Marshal(data)
	if err != nil {
		return err
	}
	filePath := fs.cacheDir + "/cache.json"
	fmt.Println("Saving cache to", filePath)
	return os.WriteFile(filePath, serialized, 0o644)
}

// LoadCache reads and deserializes the cache data from a file.
func (fs *FileStore[C, Cfg]) LoadCache() (C, error) {
	var data C
	filePath := fs.cacheDir + "/cache.json"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return data, nil
	}
	serialized, err := os.ReadFile(filePath)
	if err != nil {
		return data, err
	}
	err = json.Unmarshal(serialized, &data)
	return data, err
}

// SaveConfig serializes and saves the config data to a file.
func (fs *FileStore[C, Cfg]) SaveConfig(config Cfg) error {
	serialized, err := json.Marshal(config)
	if err != nil {
		return err
	}
	filePath := fs.configDir + "/config.json"
	return os.WriteFile(filePath, serialized, 0o644)
}

// LoadConfig reads and deserializes the config data from a file.
func (fs *FileStore[C, Cfg]) LoadConfig() (Cfg, error) {
	var config Cfg
	filePath := fs.configDir + "/config.json"
	serialized, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(serialized, &config)
	return config, err
}
