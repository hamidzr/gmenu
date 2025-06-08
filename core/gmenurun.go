package core

import (
	"fmt"
	"sync"

	"github.com/hamidzr/gmenu/model"
	"github.com/sirupsen/logrus"
)

// setExitCode sets the exit code for the application.
func (g *GMenu) SetExitCode(code model.ExitCode) error {
	logrus.Debug("setting exit code to: ", code)
	if g.ExitCode != model.Unset && g.ExitCode != code {
		msg := fmt.Sprintf("Exit code set multiple times to different values: %v -> %v", g.ExitCode, code)
		return fmt.Errorf(msg)
	}
	g.ExitCode = code
	return nil
}

// QuitWithCode exits the application.
func (g *GMenu) QuitWithCode(code model.ExitCode) {
	defer g.Quit()
	err := g.SetExitCode(code)
	if err != nil {
		panic(err)
	}
}

// Quit exits the application with the preset exit code.
func (g *GMenu) Quit() {
	if g.ExitCode == model.Unset {
		panic("Exit code not set")
	}
	// Set visibility state to false when quitting
	g.setShown(false)
	g.app.Quit()
	_ = removePidFile(g.menuID)
}

// Reset resets the gmenu state without exiting.
// Exiting and restarting is expensive.
func (g *GMenu) Reset(resetInput bool) {
	logrus.Info("resetting gmenu state")
	g.menuCancel()
	g.ui.SearchEntry.Enable()
	g.ExitCode = model.Unset
	g.SelectionWg = sync.WaitGroup{}
	if resetInput {
		g.ui.SearchEntry.SetText("")
	}
	g.menu.Selected = 0
	// Note: we don't reset isShown here as Reset doesn't change visibility
	// The visibility state should be preserved during reset
	g.SetupMenu([]string{model.LoadingItem.Title}, "")
}

func (g *GMenu) RunAppForever() error {
	if g.isRunning {
		panic("Run called multiple times")
	}
	g.isRunning = true

	// create PID file when app starts running
	_, err := createPidFile(g.menuID)
	if err != nil {
		return fmt.Errorf("failed to create PID file: %w", err)
	}

	g.app.Run()
	return nil
}

// HideUI hides the UI.
func (g *GMenu) HideUI() {
	g.ui.MainWindow.Hide()
	// Set visibility state
	g.setShown(false)
	// if err := util.MinimizeWindow(context.TODO(), g.ui.MainWindow.Title()); err != nil {
	// 	logrus.Error("Failed to minimize window:", err)
	// } else {
	// 	logrus.Debug("Minimized window", g.ui.MainWindow.Title())
	// }
}

// Run starts the application.
func (g *GMenu) CacheSelectedValue() error {
	selectedVal, err := g.SelectedValue()
	if err != nil {
		if cacheErr := g.clearCache(); cacheErr != nil {
			fmt.Println("Failed to clear cache:", cacheErr)
		}
		return err
	}
	err = g.cacheState(selectedVal.ComputedTitle())
	return err
}
