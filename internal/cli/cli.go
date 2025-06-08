package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hamidzr/gmenu/core"
	"github.com/hamidzr/gmenu/internal/config"
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
	RootCmd := &cobra.Command{
		Use:   "gmenu",
		Short: "gmenu is a fuzzy menu selector",
		RunE: func(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("invalid search method: %s", cfg.SearchMethod)
	}

	gmenu, err := core.NewGMenu(searchMethod, cfg)
	if err != nil {
		return fmt.Errorf("failed to create gmenu: %w", err)
	}

	if cfg.TerminalMode {
		return runTerminalMode(gmenu, cfg)
	}

	gmenu.SetupMenu([]string{}, cfg.InitialQuery)
	gmenu.ShowUI()
	go func() {
		items := readItems()
		if len(items) == 0 {
			logrus.Error("No items provided through standard input")
			gmenu.QuitWithCode(1)
			// signal selection is done by calling markSelectionMade instead of direct Done()
			return
		}
		gmenu.SetItems(items, nil)
	}()
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
		return err
	}
	if gmenu.GetExitCode() != model.NoError {
		logrus.Trace("Quitting gmenu with code: ", gmenu.GetExitCode())
		os.Exit(int(gmenu.GetExitCode()))
	}
	val, err := gmenu.SelectedValue()
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	// Output the selected value directly to stdout without any logging
	fmt.Println(val.ComputedTitle())
	return nil
}

func runTerminalMode(gmenu *core.GMenu, cfg *model.Config) error {
	logrus.Info("Running in terminal mode")
	items := readItems()
	if len(items) == 0 {
		logrus.Error("No items provided through standard input")
		return fmt.Errorf("no items provided through standard input")
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

	finalQuery := core.ReadUserInputLive(cfg, queryChan)
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
