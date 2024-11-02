package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/hamidzr/gmenu/core"
	"github.com/hamidzr/gmenu/internal/logger"
	"github.com/hamidzr/gmenu/model"
	"github.com/sirupsen/logrus"
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

func run() {
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

func main() {
	cmd := initCLI()
	logger.SetupLogger()
	if err := cmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
