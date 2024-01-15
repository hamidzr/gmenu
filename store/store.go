package store

import (
	"encoding/json"
	"os"
)

/*
save and load data for cache and config usage
*/

type Cache struct {
	UsageCount       map[string]int `json:"usageCount"`
	NotFoundAccepted []string       `json:"notFoundAccepted"`
	LastEntry        string         `json:"lastEntry"`
}

type Config struct {
	AppTitle      string `json:"appTitle"`
	DefaultPrompt string `json:"defaultPrompt"`
	DefaultLimit  int    `json:"defaultLimit"`
	WindowSize    struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"windowSize"`
	SearchMethod string `json:"searchMethod"`
}

type Store interface {
	SaveCache(menuID string, data Cache) error
	LoadCache(menuID string) (Cache, error)
	SaveConfig(menuID string, config Config) error
	LoadConfig(menuID string) (Config, error)
}

type FileStore struct {
	cachePath  string
	configPath string
}

func NewFileStore(cachePath, configPath string) *FileStore {
	return &FileStore{
		cachePath:  cachePath,
		configPath: configPath,
	}
}

func (fs *FileStore) SaveCache(menuID string, data Cache) error {
	// save to a single json file with menuID as key
	serialized, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(fs.cachePath, serialized, 0644)
}

func (fs *FileStore) LoadCache(menuID string) (Cache, error) {
	// load from a single json file with menuID as key
	var data Cache
	serialized, err := os.ReadFile(fs.cachePath)
	if err != nil {
		return data, err
	}
	err = json.Unmarshal(serialized, &data)
	return data, err
}

func (fs *FileStore) SaveConfig(menuID string, config Config) error {
	// save to a single json file with menuID as key
	serialized, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(fs.configPath, serialized, 0644)
}

func (fs *FileStore) LoadConfig(menuID string) (Config, error) {
	// load from a single json file with menuID as key
	var config Config
	serialized, err := os.ReadFile(fs.configPath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(serialized, &config)
	return config, err
}
