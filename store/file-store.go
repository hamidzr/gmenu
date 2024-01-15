package store

import (
	"encoding/json"
	"fmt"
	"os"
)

type FileStore struct {
	cacheDir  string
	configDir string
}

var unsetMenuIDErr = fmt.Errorf("menuID cannot be empty")

func NewFileStore(cacheDir, configDir string) (*FileStore, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}
	return &FileStore{
		cacheDir:  cacheDir,
		configDir: configDir,
	}, nil
}

func (fs *FileStore) SaveCache(menuID string, data Cache) error {
	serialized, err := json.Marshal(data)
	if err != nil {
		return err
	}
	filePath := fs.cacheDir + "/" + menuID + ".json"
	return os.WriteFile(filePath, serialized, 0644)
}

func (fs *FileStore) LoadCache(menuID string) (Cache, error) {
	var data Cache
	filePath := fs.cacheDir + "/" + menuID + ".json"
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

func (fs *FileStore) SaveConfig(menuID string, config Config) error {
	serialized, err := json.Marshal(config)
	if err != nil {
		return err
	}
	filePath := fs.configDir + "/" + menuID + ".json"
	return os.WriteFile(filePath, serialized, 0644)
}

func (fs *FileStore) LoadConfig(menuID string) (Config, error) {
	var config Config
	filePath := fs.configDir + "/" + menuID + ".json"
	serialized, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(serialized, &config)
	return config, err
}
