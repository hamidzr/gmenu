package core

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hamidzr/gmenu/model"
	"golang.org/x/term"
)

// ReadUserInputLive reads user input live from the terminal
// keeps a local repr of the text user put in and maintaint a line of output
// that shows the user's input so far
func ReadUserInputLive(cfg *model.Config, queryChan chan<- string) string {
	// Set up raw terminal mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Failed to set raw terminal mode:", err)
		return ""
	}
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			fmt.Printf("Failed to restore terminal: %v\n", err)
		}
	}()

	// Create a new reader from stdin
	reader := bufio.NewReader(os.Stdin)
	input := []byte(cfg.InitialQuery)

	// Display initial prompt with initial query
	fmt.Printf("\r%s%s", cfg.Prompt, string(input))
	// Send initial query
	queryChan <- string(input)

	// Set up channel for handling signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Create a channel to signal when Enter is pressed
	enterPressed := make(chan bool, 1)
	inputCh := make(chan string, 1)

	// Start goroutine to read input
	go func() {
		for {
			char, err := reader.ReadByte()
			if err != nil {
				fmt.Println("\nError reading input:", err)
				enterPressed <- true
				return
			}

			if char == '\r' || char == '\n' {
				// Enter pressed
				inputCh <- string(input)
				enterPressed <- true
				return
			} else if char == 127 || char == 8 {
				// Backspace/Delete pressed
				if len(input) > 0 {
					input = input[:len(input)-1]
					queryChan <- string(input)
				}
			} else if char == 3 {
				// Ctrl+C pressed
				fmt.Printf("\n%sInput cancelled\n", cfg.Prompt)
				inputCh <- ""
				enterPressed <- true
				return
			} else if char >= 32 && char <= 126 {
				// Printable ASCII characters
				input = append(input, char)
				queryChan <- string(input)
			}
		}
	}()

	// Wait for either a signal or enter press
	select {
	case <-sigCh:
		fmt.Printf("\n%sInput cancelled\n", cfg.Prompt)
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			fmt.Printf("Failed to restore terminal: %v\n", err)
		}
		os.Exit(0) // Exit immediately on signal
	case <-enterPressed:
		fmt.Println()
		return <-inputCh
	}
	return ""
}
