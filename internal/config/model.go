package config

import (
	"strings"

	"github.com/hamidzr/gmenu/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// BindFlags binds CLI flags to the cobra command
func BindFlags(cmd *cobra.Command) {
	defaults := model.DefaultConfig()

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
	defaults := model.DefaultConfig()
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
