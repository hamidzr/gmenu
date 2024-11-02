package main

import (
	"github.com/hamidzr/gmenu/constant"
	"github.com/hamidzr/gmenu/model"
	"github.com/spf13/cobra"
)

var cliArgs = model.CliArgs{
	Title:         constant.ProjectName,
	Prompt:        "Search",
	MenuID:        "",
	SearchMethod:  "fuzzy",
	PreserveOrder: false,
	InitialQuery:  "",
	AutoAccept:    false,
}

func initCLI() *cobra.Command {
	RootCmd := &cobra.Command{
		Use:   "gmenu",
		Short: "gmenu is a fuzzy menu selector",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}

	RootCmd.PersistentFlags().StringVarP(&cliArgs.Title, "title", "t", cliArgs.Title, "Title of the menu window")
	RootCmd.PersistentFlags().StringVarP(&cliArgs.InitialQuery, "initial-query", "q", cliArgs.InitialQuery, "Initial query to search for")
	RootCmd.PersistentFlags().StringVarP(&cliArgs.Prompt, "prompt", "p", cliArgs.Prompt, "Prompt of the menu window")
	RootCmd.PersistentFlags().StringVarP(&cliArgs.MenuID, "menu-id", "m", cliArgs.MenuID, "Menu ID")
	RootCmd.PersistentFlags().StringVarP(&cliArgs.SearchMethod, "search-method", "s", cliArgs.SearchMethod, "Search method")
	RootCmd.PersistentFlags().BoolVarP(&cliArgs.PreserveOrder, "preserve-order", "o", cliArgs.PreserveOrder, "Preserve the order of the input items")
	RootCmd.PersistentFlags().BoolVarP(&cliArgs.AutoAccept, "auto-accept", "", cliArgs.AutoAccept, "Auto accept if there's only a single match.")

	return RootCmd
}
