package cli

import (
	"bufio"
	"fmt"
	"os"

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
		// gmenu.AttemptAutoSelect()
	}()
	go func() {
		// if selection is made without an exit, stop the app.
		gmenu.SelectionWg.Wait()
		if gmenu.ExitCode == model.Unset {
			gmenu.QuitWithCode(0)
		} else {
			gmenu.QuitWithCode(gmenu.ExitCode)
		}
	}()

	// go func() {
	// 	for {
	// 		gmenu.ShowUI()
	// 		time.Sleep(2 * time.Second)
	// 		go gmenu.CacheSelectedValue()
	// 		item, err := gmenu.SelectedValue()
	// 		if	err == nil {
	// 			fmt.Println(item.ComputedTitle())
	// 		}
	// 		gmenu.ToggleVisibility()
	// 		gmenu.Reset()
	// 	}
	// }()

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
	fmt.Println(val.ComputedTitle())
}
