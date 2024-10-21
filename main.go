package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/hamidzr/gmenu/constant"
	"github.com/hamidzr/gmenu/core"
	"github.com/hamidzr/gmenu/internal/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// CliArgs is a struct to hold the root CLI arguments.
type CliArgs struct {
	// Title string
	title string
	// Menu prompt string
	prompt string
	// Menu ID
	menuID string
	// Search method
	searchMethod string
	// Preserve the order of the input items.
	preserveOrder bool
	// initial query
	initialQuery string
	// TODO: Allow custom output.
	// allowCustomOutput bool
}

var cliArgs = CliArgs{
	title:         constant.ProjectName,
	prompt:        "Search",
	menuID:        "",
	searchMethod:  "fuzzy",
	preserveOrder: false,
	initialQuery:  "",
	// allowCustomOutput: true,
}

func initCLI() *cobra.Command {
	RootCmd := &cobra.Command{
		Use:   "gmenu",
		Short: "gmenu is a fuzzy menu selector",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}

	RootCmd.PersistentFlags().StringVarP(&cliArgs.title, "title", "t", cliArgs.title, "Title of the menu window")
	RootCmd.PersistentFlags().StringVarP(&cliArgs.initialQuery, "initial-query", "q", cliArgs.initialQuery, "Initial query to search for")
	RootCmd.PersistentFlags().StringVarP(&cliArgs.prompt, "prompt", "p", cliArgs.prompt, "Prompt of the menu window")
	RootCmd.PersistentFlags().StringVarP(&cliArgs.menuID, "menu-id", "m", cliArgs.menuID, "Menu ID")
	RootCmd.PersistentFlags().StringVarP(&cliArgs.searchMethod, "search-method", "s", cliArgs.searchMethod, "Search method")
	RootCmd.PersistentFlags().BoolVarP(&cliArgs.preserveOrder, "preserve-order", "o", cliArgs.preserveOrder, "Preserve the order of the input items")

	return RootCmd
}

func readItems() []string {
	var items []string
	info, _ := os.Stdin.Stat()
	if (info.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			items = append(items, line)
		}
		if err := scanner.Err(); err != nil {
			logrus.Error("error reading standard input:", err)
			os.Exit(1)
		}
	}
	return items
}

func run() {
	searchMethod, ok := core.SearchMethods[cliArgs.searchMethod]
	if !ok {
		logrus.Error("Invalid search method")
		os.Exit(1)
	}
	gmenu, err := core.NewGMenu(cliArgs.title, cliArgs.prompt, nil, cliArgs.menuID)
	if err != nil {
		logrus.Error(err, "failed to create gmenu")
		os.Exit(1)
	}
	gmenu.SetupMenu([]string{"Loading"}, cliArgs.initialQuery, searchMethod, cliArgs.preserveOrder)
	gmenu.SetupUI()
	go func() {
		items := readItems()
		if len(items) == 0 {
			logrus.Error("No items provided through standard input")
			gmenu.Quit(1)
			return
		}
		gmenu.SetItems(items, nil)
	}()
	if err := gmenu.Run(); err != nil {
		logrus.WithError(err).Error("run() err")
	}
	if gmenu.ExitCode != 0 {
		logrus.Trace("Quitting gmenu with code: ", gmenu.ExitCode)
		os.Exit(gmenu.ExitCode)
	}
	val, err := gmenu.SelectedValue()
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	fmt.Println(val.ComputedTitle())
}

func main() {
	cmd := initCLI()
	logger.SetupLogger()
	if err := cmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
