package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hamidzr/gmenu/constant"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Config holds all configuration for the application
type Config struct {
	// app settings
	Title              string  `mapstructure:"title" yaml:"title"`
	Prompt             string  `mapstructure:"prompt" yaml:"prompt"`
	MenuID             string  `mapstructure:"menu_id" yaml:"menu_id"`
	SearchMethod       string  `mapstructure:"search_method" yaml:"search_method"`
	PreserveOrder      bool    `mapstructure:"preserve_order" yaml:"preserve_order"`
	InitialQuery       string  `mapstructure:"initial_query" yaml:"initial_query"`
	AutoAccept         bool    `mapstructure:"auto_accept" yaml:"auto_accept"`
	TerminalMode       bool    `mapstructure:"terminal_mode" yaml:"terminal_mode"`
	NoNumericSelection bool    `mapstructure:"no_numeric_selection" yaml:"no_numeric_selection"`
	MinWidth           float32 `mapstructure:"min_width" yaml:"min_width"`
	MinHeight          float32 `mapstructure:"min_height" yaml:"min_height"`
	MaxWidth           float32 `mapstructure:"max_width" yaml:"max_width"`
	MaxHeight          float32 `mapstructure:"max_height" yaml:"max_height"`

	// internal settings
	AcceptCustomSelection bool `mapstructure:"accept_custom_selection" yaml:"accept_custom_selection"`
}

// ConfigWithComments represents the config structure with YAML comments for file generation
type ConfigWithComments struct {
	Config `yaml:",inline"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		Title:                 constant.ProjectName,
		Prompt:                "Search",
		MenuID:                "",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "",
		AutoAccept:            false,
		TerminalMode:          false,
		NoNumericSelection:    false,
		MinWidth:              600,
		MinHeight:             300,
		MaxWidth:              0, // auto-calculated
		MaxHeight:             0, // auto-calculated
		AcceptCustomSelection: true,
	}
}

// getConfigPaths returns the config directory paths in priority order
func getConfigPaths(menuID string) []string {
	var paths []string

	// when menu ID is provided, prioritize namespaced configs
	if menuID != "" {
		if configDir, err := os.UserConfigDir(); err == nil {
			paths = append(paths, filepath.Join(configDir, "gmenu", menuID))
		}
		if homeDir, err := os.UserHomeDir(); err == nil {
			paths = append(paths, filepath.Join(homeDir, ".config", "gmenu", menuID))
			paths = append(paths, filepath.Join(homeDir, ".gmenu", menuID))
		}
	}

	// add default config paths
	if configDir, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(configDir, "gmenu"))
	}
	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(homeDir, ".config", "gmenu"))
		paths = append(paths, filepath.Join(homeDir, ".gmenu"))
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
func InitConfig(cmd *cobra.Command) (*Config, error) {
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
	v.SetEnvPrefix("GMENU")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	// set defaults
	defaults := DefaultConfig()
	v.SetDefault("title", defaults.Title)
	v.SetDefault("prompt", defaults.Prompt)
	v.SetDefault("menu_id", defaults.MenuID)
	v.SetDefault("search_method", defaults.SearchMethod)
	v.SetDefault("preserve_order", defaults.PreserveOrder)
	v.SetDefault("initial_query", defaults.InitialQuery)
	v.SetDefault("auto_accept", defaults.AutoAccept)
	v.SetDefault("terminal_mode", defaults.TerminalMode)
	v.SetDefault("no_numeric_selection", defaults.NoNumericSelection)
	v.SetDefault("min_width", defaults.MinWidth)
	v.SetDefault("min_height", defaults.MinHeight)
	v.SetDefault("max_width", defaults.MaxWidth)
	v.SetDefault("max_height", defaults.MaxHeight)
	v.SetDefault("accept_custom_selection", defaults.AcceptCustomSelection)

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
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// BindFlags binds CLI flags to the cobra command
func BindFlags(cmd *cobra.Command) {
	defaults := DefaultConfig()

	cmd.PersistentFlags().StringP("title", "t", defaults.Title, "Title of the menu window")
	cmd.PersistentFlags().StringP("initial-query", "q", defaults.InitialQuery, "Initial query to search for")
	cmd.PersistentFlags().StringP("prompt", "p", defaults.Prompt, "Prompt of the menu window")
	cmd.PersistentFlags().StringP("menu-id", "m", defaults.MenuID, "Menu ID")
	cmd.PersistentFlags().StringP("search-method", "s", defaults.SearchMethod, "Search method")
	cmd.PersistentFlags().BoolP("preserve-order", "o", defaults.PreserveOrder, "Preserve the order of the input items")
	cmd.PersistentFlags().Bool("auto-accept", defaults.AutoAccept, "Auto accept if there's only a single match")
	cmd.PersistentFlags().Bool("terminal", defaults.TerminalMode, "Run in terminal-only mode without GUI")
	cmd.PersistentFlags().Bool("no-numeric-selection", defaults.NoNumericSelection, "Disable numeric selection")
	cmd.PersistentFlags().Float32("min-width", defaults.MinWidth, "Minimum window width")
	cmd.PersistentFlags().Float32("min-height", defaults.MinHeight, "Minimum window height")
	cmd.PersistentFlags().Float32("max-width", defaults.MaxWidth, "Maximum window width (0 for auto-calculated)")
	cmd.PersistentFlags().Float32("max-height", defaults.MaxHeight, "Maximum window height (0 for auto-calculated)")
	cmd.PersistentFlags().Bool("init-config", false, "Generate and save default config file")
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
	defaults := DefaultConfig()
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
