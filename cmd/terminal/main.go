package main

import (
	"fmt"
	"strings"

	"github.com/hamidzr/gmenu/core"
	"github.com/hamidzr/gmenu/model"
)

func main() {
	cfg := &model.Config{
		Prompt:       "Enter text: ",
		InitialQuery: "Hello world",
	}

	// Create channel for receiving query updates
	queryChan := make(chan string, 1)

	// Start a goroutine to process query updates
	go func() {
		for query := range queryChan {
			// Clear the line below the input
			fmt.Printf("\n\r%s\r", strings.Repeat(" ", 80))
			// Print whether query length is even or odd
			if len(query)%2 == 0 {
				fmt.Printf("\n\rQuery length is even: %d", len(query))
			} else {
				fmt.Printf("\n\rQuery length is odd: %d", len(query))
			}
			// Move cursor back up to input line
			fmt.Printf("\033[1A\r%s%s", cfg.Prompt, query)
		}
	}()

	result, err := core.ReadUserInputLive(cfg, queryChan)
	if err != nil {
		fmt.Printf("\nInput ended: %v\n", err)
		return
	}
	if result != "" {
		fmt.Printf("\nFinal input: %s\n", result)
	}
}
