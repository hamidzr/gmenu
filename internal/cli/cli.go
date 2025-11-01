package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hamidzr/gmenu/core"
	"github.com/hamidzr/gmenu/internal/config"
	"github.com/hamidzr/gmenu/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func readItems() ([]string, error) {
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
			return nil, fmt.Errorf("error reading standard input: %w", err)
		}
	}
	return items, nil
}

func InitCLI() *cobra.Command {
	RootCmd := &cobra.Command{
		Use:           "gmenu",
		Short:         "gmenu is a fuzzy menu selector",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				msg := fmt.Sprintf("unknown argument(s): %s", strings.Join(args, " "))
				cmd.PrintErrln(msg)
				return model.NewExitError(model.UnknownError, errors.New(msg))
			}
			// check if user wants to initialize config
			initConfig, _ := cmd.Flags().GetBool("init-config")
			if initConfig {
				menuID, _ := cmd.Flags().GetString("menu-id")
				configPath, err := config.InitConfigFile(menuID)
				if err != nil {
					return fmt.Errorf("failed to initialize config: %w", err)
				}
				fmt.Printf("✅ Config file created successfully at: %s\n", configPath)
				if menuID != "" {
					fmt.Printf("📁 Menu ID: %s\n", menuID)
					fmt.Printf("💡 Use with: gmenu --menu-id %s\n", menuID)
				} else {
					fmt.Printf("💡 This is the default config file\n")
				}
				fmt.Printf("📝 Edit the file to customize your settings\n")
				return nil
			}

			// initialize configuration with proper priority handling
			cfg, err := config.InitConfig(cmd)
			if err != nil {
				fmt.Println("err", err)
				return fmt.Errorf("failed to initialize config: %w", err)
			}

			return run(cfg)
		},
	}

	// bind all flags using the new config system
	config.BindFlags(RootCmd)

	return RootCmd
}

func run(cfg *model.Config) error {
	searchMethod, ok := core.SearchMethods[cfg.SearchMethod]
	if !ok {
		return model.NewExitError(model.UnknownError, fmt.Errorf("invalid search method: %s", cfg.SearchMethod))
	}

	gmenu, err := core.NewGMenu(searchMethod, cfg)
	if err != nil {
		return model.NewExitError(model.UnknownError, fmt.Errorf("failed to create gmenu: %w", err))
	}

	if cfg.TerminalMode {
		return runTerminalMode(gmenu, cfg)
	}

	items, err := readItems()
	if err != nil {
		return model.NewExitError(model.UnknownError, err)
	}
	if len(items) == 0 {
		logrus.Error("No items provided through standard input")
		gmenu.QuitWithCode(model.UnknownError)
		return model.NewExitError(model.UnknownError, fmt.Errorf("no items provided through standard input"))
	}

	if err := gmenu.SetupMenu(items, cfg.InitialQuery); err != nil {
		return model.NewExitError(model.UnknownError, fmt.Errorf("failed to setup menu: %w", err))
	}

	if cfg.AutoAccept {
		if gmenu.AttemptAutoSelect() {
			val, err := gmenu.SelectedValue()
			if err != nil {
				return model.NewExitError(model.UnknownError, fmt.Errorf("auto-select failed to retrieve value: %w", err))
			}
			fmt.Println(val.ComputedTitle())
			return nil
		}
		logrus.WithField("matches", gmenu.MatchCount()).
			Debug("auto-accept conditions not met; falling back to interactive mode")
	}

	if err := gmenu.ShowUI(); err != nil {
		return fmt.Errorf("failed to show UI: %w", err)
	}
	go func() {
		gmenu.WaitForSelection()
		if gmenu.GetExitCode() == model.Unset {
			gmenu.QuitWithCode(0)
		} else {
			gmenu.Quit()
		}
	}()

	if err := gmenu.RunAppForever(); err != nil {
		logrus.WithError(err).Error("run() err")
		return model.NewExitError(model.UnknownError, err)
	}
	if gmenu.GetExitCode() != model.NoError {
		logrus.Trace("Quitting gmenu with code: ", gmenu.GetExitCode())
		return model.NewExitError(gmenu.GetExitCode(), nil)
	}
	val, err := gmenu.SelectedValue()
	if err != nil {
		logrus.Error(err)
		return model.NewExitError(model.UnknownError, err)
	}
	// Output the selected value directly to stdout without any logging
	fmt.Println(val.ComputedTitle())
	return nil
}

func runTerminalMode(gmenu *core.GMenu, cfg *model.Config) error {
	logrus.Info("Running in terminal mode")
	items, err := readItems()
	if err != nil {
		return model.NewExitError(model.UnknownError, err)
	}
	if len(items) == 0 {
		logrus.Error("No items provided through standard input")
		return model.NewExitError(model.UnknownError, fmt.Errorf("no items provided through standard input"))
	}
	// reset stdin from non-interactive to interactive.
	_ = os.Stdin.Close()
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
		// ReadUserInputLive() will close queryChan when the user is done.
		for query := range queryChan {
			// Clear screen and reset cursor
			fmt.Print("\033[2J\033[H")

			// Print header
			logrus.Infof("%s: %s", cfg.Prompt, query)
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

	finalQuery, inputErr := core.ReadUserInputLive(cfg, queryChan)
	if inputErr != nil {
		switch {
		case errors.Is(inputErr, core.ErrTerminalInterrupted):
			return model.NewExitError(model.NoError, nil)
		case errors.Is(inputErr, core.ErrTerminalCancelled):
			return model.NewExitError(model.UserCanceled, nil)
		default:
			return model.NewExitError(model.UnknownError, inputErr)
		}
	}

	matches := matcher(items, finalQuery)
	if len(matches) == 0 {
		logrus.Info("No matches found")
		return nil
	}
	if len(matches) > 1 {
		logrus.Info("Multiple matches found. Picking the first one.")
	}
	// clear the screen and show result
	fmt.Print("\033[2J\033[H")
	// Output the selected value directly to stdout without any logging
	fmt.Println(matches[0])
	return nil
}
