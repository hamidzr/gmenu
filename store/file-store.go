package store

import (
	"encoding/json"
	"os"
)

type FileStore struct {
	cacheDir  string
	configDir string
}

func NewFileStore(namespace string) (*FileStore, error) {
	cacheDir := CacheDir(namespace)
	configDir := ConfigDir(namespace)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return nil, err
	}
	return &FileStore{
		cacheDir:  cacheDir,
		configDir: configDir,
	}, nil
}

func (fs *FileStore) SaveCache(data Cache) error {
	serialized, err := json.Marshal(data)
	if err != nil {
		return err
	}
	filePath := fs.cacheDir + "/cache.json"
	return os.WriteFile(filePath, serialized, 0o644)
}

func (fs *FileStore) LoadCache() (Cache, error) {
	var data Cache
	filePath := fs.cacheDir + "/cache.json"
	// if file doesn't exist return empty cache
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

func (fs *FileStore) SaveConfig(config Config) error {
	serialized, err := json.Marshal(config)
	if err != nil {
		return err
	}
	filePath := fs.configDir + "/config.json"
	return os.WriteFile(filePath, serialized, 0o644)
}

func (fs *FileStore) LoadConfig() (Config, error) {
	var config Config
	filePath := fs.configDir + "/config.json"
	serialized, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(serialized, &config)
	return config, err
}
