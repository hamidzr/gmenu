package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hamidzr/gmenu/model"
	"gopkg.in/yaml.v2"
)

// getConfigPaths returns the config directory paths in priority order
// prefers ~/.config over macos application support dir
func GetConfigPaths(menuID string) []string {
	var paths []string

	// when menu ID is provided, prioritize namespaced configs
	if menuID != "" {
		if homeDir, err := os.UserHomeDir(); err == nil {
			paths = append(paths, filepath.Join(homeDir, ".config", "gmenu", menuID))
			paths = append(paths, filepath.Join(homeDir, ".gmenu", menuID))
		}
		if configDir, err := os.UserConfigDir(); err == nil {
			paths = append(paths, filepath.Join(configDir, "gmenu", menuID))
		}
	}

	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(homeDir, ".config", "gmenu"))
		paths = append(paths, filepath.Join(homeDir, ".gmenu"))
	}
	if configDir, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(configDir, "gmenu"))
	}
	paths = append(paths, ".")

	return paths
}

// getPreferredConfigDir returns the preferred config directory for writing
func GetPreferredConfigDir(menuID string) (string, error) {
	if homeDir, err := os.UserHomeDir(); err == nil {
		if menuID != "" {
			return filepath.Join(homeDir, ".config", "gmenu", menuID), nil
		}
		return filepath.Join(homeDir, ".config", "gmenu"), nil
	}

	if userConfigDir, err := os.UserConfigDir(); err == nil {
		if menuID != "" {
			return filepath.Join(userConfigDir, "gmenu", menuID), nil
		}
		return filepath.Join(userConfigDir, "gmenu"), nil
	}

	return "", fmt.Errorf("unable to determine config directory")
}

// GetConfigByMenuID loads the config for a given menu ID from the filesystem.
// It searches for the first config.yaml file in the config paths for the menu ID.
func GetConfigByMenuID(menuID string) (*model.Config, error) {
	paths := GetConfigPaths(menuID)
	for _, dir := range paths {
		configPath := filepath.Join(dir, "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				return nil, err
			}
			cfg := model.DefaultConfig()
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, err
			}
			return cfg, nil
		}
	}
	return nil, fmt.Errorf("no config.yaml found for menu id '%s'", menuID)
}
