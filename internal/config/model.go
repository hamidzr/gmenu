package config

import (
	"strings"

	"github.com/hamidzr/gmenu/constant"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

// SetViperDefaults sets default values in viper configuration
func SetViperDefaults(v *viper.Viper) {
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
}

// SetViperEnvSettings configures viper environment variable settings
func SetViperEnvSettings(v *viper.Viper) {
	v.SetEnvPrefix("GMENU")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
}
