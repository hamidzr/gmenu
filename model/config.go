package model

import "github.com/hamidzr/gmenu/constant"

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
