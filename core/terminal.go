package core

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hamidzr/gmenu/model"
	"golang.org/x/term"
)

var (
	// ErrTerminalInterrupted indicates the input loop was interrupted by a signal.
	ErrTerminalInterrupted = errors.New("terminal input interrupted")
	// ErrTerminalCancelled indicates the user cancelled input (e.g. via Ctrl+C).
	ErrTerminalCancelled = errors.New("terminal input cancelled")
)

// ReadUserInputLive reads user input live from the terminal
// keeps a local repr of the text user put in and maintaint a line of output
// that shows the user's input so far
func ReadUserInputLive(cfg *model.Config, queryChan chan<- string) (string, error) {
	// Set up raw terminal mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Failed to set raw terminal mode:", err)
		return "", err
	}
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			fmt.Printf("Failed to restore terminal: %v\n", err)
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	input := []byte(cfg.InitialQuery)

	fmt.Printf("\r%s%s", cfg.Prompt, string(input))
	queryChan <- string(input)
	defer close(queryChan)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	enterPressed := make(chan struct{}, 1)
	resultCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		for {
			char, err := reader.ReadByte()
			if err != nil {
				fmt.Println("\nError reading input:", err)
				select {
				case errCh <- fmt.Errorf("reading input: %w", err):
				default:
				}
				select {
				case enterPressed <- struct{}{}:
				default:
				}
				return
			}

			switch {
			case char == '\r' || char == '\n':
				resultCh <- string(input)
				select {
				case enterPressed <- struct{}{}:
				default:
				}
				return
			case char == 127 || char == 8:
				if len(input) > 0 {
					input = input[:len(input)-1]
					queryChan <- string(input)
				}
			case char == 3:
				fmt.Printf("\n%sInput cancelled\n", cfg.Prompt)
				select {
				case errCh <- ErrTerminalCancelled:
				default:
				}
				select {
				case enterPressed <- struct{}{}:
				default:
				}
				return
			case char >= 32 && char <= 126:
				input = append(input, char)
				queryChan <- string(input)
			}
		}
	}()

	select {
	case sig := <-sigCh:
		fmt.Printf("\n%sInput cancelled (%s)\n", cfg.Prompt, sig.String())
		return "", ErrTerminalInterrupted
	case <-enterPressed:
		fmt.Println()
		select {
		case err := <-errCh:
			return "", err
		case res := <-resultCh:
			return res, nil
		default:
			return "", nil
		}
	}
}
