package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hamidzr/gmenu/constant"
	"github.com/hamidzr/gmenu/core"
	"github.com/hamidzr/gmenu/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

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

func InitCLI() *cobra.Command {
	var cliArgs = model.CliArgs{
		Title:         constant.ProjectName,
		Prompt:        "Search",
		MenuID:        "",
		SearchMethod:  "fuzzy",
		PreserveOrder: false,
		InitialQuery:  "",
		AutoAccept:    false,
		TerminalMode:  false,
	}

	RootCmd := &cobra.Command{
		Use:   "gmenu",
		Short: "gmenu is a fuzzy menu selector",
		Run: func(cmd *cobra.Command, args []string) {
			run(cliArgs)
		},
	}

	RootCmd.PersistentFlags().StringVarP(&cliArgs.Title, "title", "t", cliArgs.Title, "Title of the menu window")
	RootCmd.PersistentFlags().StringVarP(&cliArgs.InitialQuery, "initial-query", "q", cliArgs.InitialQuery, "Initial query to search for")
	RootCmd.PersistentFlags().StringVarP(&cliArgs.Prompt, "prompt", "p", cliArgs.Prompt, "Prompt of the menu window")
	RootCmd.PersistentFlags().StringVarP(&cliArgs.MenuID, "menu-id", "m", cliArgs.MenuID, "Menu ID")
	RootCmd.PersistentFlags().StringVarP(&cliArgs.SearchMethod, "search-method", "s", cliArgs.SearchMethod, "Search method")
	RootCmd.PersistentFlags().BoolVarP(&cliArgs.PreserveOrder, "preserve-order", "o", cliArgs.PreserveOrder, "Preserve the order of the input items")
	RootCmd.PersistentFlags().BoolVarP(&cliArgs.AutoAccept, "auto-accept", "", cliArgs.AutoAccept, "Auto accept if there's only a single match.")
	RootCmd.PersistentFlags().BoolVarP(&cliArgs.TerminalMode, "terminal", "", cliArgs.TerminalMode, "Run in terminal-only mode without GUI")

	return RootCmd
}

func run(cliArgs model.CliArgs) {
	searchMethod, ok := core.SearchMethods[cliArgs.SearchMethod]
	if !ok {
		logrus.Error("Invalid search method")
		os.Exit(1)
	}
	conf := model.DefaultConfig()
	conf.CliArgs = cliArgs
	gmenu, err := core.NewGMenu(searchMethod, conf)
	if err != nil {
		logrus.Error(err, "failed to create gmenu")
		os.Exit(1)
	}

	if cliArgs.TerminalMode {
		runTerminalMode(gmenu, cliArgs)
		return
	}

	gmenu.SetupMenu([]string{"Loading"}, cliArgs.InitialQuery)
	gmenu.ShowUI()
	go func() {
		items := readItems()
		if len(items) == 0 {
			logrus.Error("No items provided through standard input")
			gmenu.QuitWithCode(1)
			gmenu.SelectionWg.Done()
			return
		}
		gmenu.SetItems(items, nil)
	}()
	go func() {
		gmenu.SelectionWg.Wait()
		if gmenu.ExitCode == model.Unset {
			gmenu.QuitWithCode(0)
		} else {
			gmenu.QuitWithCode(gmenu.ExitCode)
		}
	}()

	if err := gmenu.RunAppForever(); err != nil {
		logrus.WithError(err).Error("run() err")
	}
	if gmenu.ExitCode != model.NoError {
		logrus.Trace("Quitting gmenu with code: ", gmenu.ExitCode)
		os.Exit(int(gmenu.ExitCode))
	}
	val, err := gmenu.SelectedValue()
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	// Output the selected value directly to stdout without any logging
	fmt.Println(val.ComputedTitle())
}

func runTerminalMode(gmenu *core.GMenu, cliArgs model.CliArgs) {
	logrus.Info("Running in terminal mode")
	items := readItems()
	if len(items) == 0 {
		logrus.Error("No items provided through standard input")
		return
	}
	// reset stdin from non-interactive to interactive.
	os.Stdin.Close()
	os.Stdin, _ = os.Open("/dev/tty")

	matcher := func(items []string, query string) []string {
		var matches []string
		for _, item := range items {
			if strings.Contains(strings.ToLower(item), strings.ToLower(query)) {
				matches = append(matches, item)
			}
		}
		return matches
	}

	queryChan := make(chan string, 1)
	go func() {
		// ReadUserInputLive() will close the queryChan when the user is done.
		for query := range queryChan {
			// Clear screen and reset cursor
			fmt.Print("\033[2J\033[H")

			// Print header
			logrus.Infof("%s: %s", cliArgs.Prompt, query)
			logrus.Info("--------------------------------")

			// Filter and display matching items
			matchCount := 0
			for idx, match := range matcher(items, query) {
				logrus.Infof("%d. %s", idx+1, match)
				matchCount++
			}

			if matchCount == 0 {
				logrus.Info("(no matches)")
			}
			logrus.Info("--------------------------------")
		}
	}()
	finalQuery := core.ReadUserInputLive(cliArgs, queryChan)
	matches := matcher(items, finalQuery)
	if len(matches) == 0 {
		logrus.Info("No matches found")
		return
	}
	if len(matches) > 1 {
		logrus.Info("Multiple matches found. Picking the first one.")
	}
	// clear the screen and show result
	fmt.Print("\033[2J\033[H")
	// Output the selected value directly to stdout without any logging
	fmt.Println(matches[0])
}
