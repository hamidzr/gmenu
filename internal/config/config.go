package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hamidzr/gmenu/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// ConfigWithComments represents the config structure with YAML comments for file generation
type ConfigWithComments struct {
	model.Config `yaml:",inline"`
}

// getConfigPaths returns the config directory paths in priority order
// prefers ~/.config over macos application support dir
func getConfigPaths(menuID string) []string {
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
func getPreferredConfigDir(menuID string) (string, error) {
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

// InitConfig initializes Viper configuration with proper priority:
// 1. CLI flags (highest priority)
// 2. Environment variables
// 3. Config file (lowest priority)
func InitConfig(cmd *cobra.Command) (*model.Config, error) {
	v := viper.New()

	// set config file settings - look for config.yaml to avoid conflicts with cache files
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// get menu ID from flags to determine config namespace
	menuID, _ := cmd.Flags().GetString("menu-id")

	// add config paths in priority order
	for _, path := range getConfigPaths(menuID) {
		v.AddConfigPath(path)
	}

	// set environment variable settings
	SetViperEnvSettings(v)

	// set defaults
	SetViperDefaults(v)

	// read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// config file not found is ok, we'll use defaults + env vars + flags
	}

	// bind CLI flags to viper (highest priority)
	if err := v.BindPFlags(cmd.Flags()); err != nil {
		return nil, fmt.Errorf("error binding flags: %w", err)
	}

	// unmarshal into config struct
	var config model.Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// InitConfigFile generates and saves a default config file to the appropriate location
func InitConfigFile(menuID string) (string, error) {
	// get the preferred config directory
	configDir, err := getPreferredConfigDir(menuID)
	if err != nil {
		return "", err
	}

	// create the directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	configPath := filepath.Join(configDir, "config.yaml")

	// check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		return "", fmt.Errorf("config file already exists at %s", configPath)
	}

	// create default config with the provided menu ID
	defaults := model.DefaultConfig()
	defaults.MenuID = menuID // set the menu ID in the config

	// marshal to YAML
	yamlData, err := yaml.Marshal(defaults)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// add header comment
	header := `# gmenu configuration file
# Generated automatically - customize as needed
# 
# Search method options: fuzzy, exact, regex
# Window dimensions: set to 0 for auto-calculated max dimensions
#

`

	finalContent := header + string(yamlData)

	// write the config file
	if err := os.WriteFile(configPath, []byte(finalContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write config file %s: %w", configPath, err)
	}

	return configPath, nil
}
