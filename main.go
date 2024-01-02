package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/hamidzr/gmenu/core"
	"github.com/spf13/cobra"
)

func initCLI() *cobra.Command {
	var RootCmd = &cobra.Command{
		Use:   "gmenu",
		Short: "gmenu is a fuzzy menu selector",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}

	return RootCmd
}

func readItems() []string {
	var items []string
	// Check if there is any input from stdin (piped text)
	info, _ := os.Stdin.Stat()
	if (info.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			items = append(items, line)
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "error reading standard input:", err)
			os.Exit(1)
		}
	}

	// Proceed only if there are items
	if len(items) == 0 {
		fmt.Fprintln(os.Stderr, "No items provided through standard input")
		os.Exit(1)
	}
	return items
}

func run() {
	items := readItems()
	gmenu := core.NewGMenu(items)
	gmenu.Run()
}

func main() {
	cmd := initCLI()
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
