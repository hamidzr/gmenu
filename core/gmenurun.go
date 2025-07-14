package core

import (
	"fmt"

	"github.com/frostbyte73/core"
	"github.com/hamidzr/gmenu/model"
	"github.com/sirupsen/logrus"
)

// setExitCode sets the exit code for the application.
func (g *GMenu) SetExitCode(code model.ExitCode) error {
	logrus.Debug("setting exit code to: ", code)
	if g.exitCode != model.Unset && g.exitCode != code {
		// Log the conflict but don't panic - just use the first exit code set
		logrus.Warn("Exit code already set", "current", g.exitCode, "attempted", code)
		return nil
	}
	g.exitCode = code
	return nil
}

// QuitWithCode exits the application.
func (g *GMenu) QuitWithCode(code model.ExitCode) {
	defer g.Quit()
	err := g.SetExitCode(code)
	if err != nil {
		fmt.Println("Error setting exit code:", err)
		panic(err)
	}
}

// Quit exits the application with the preset exit code.
func (g *GMenu) Quit() {
	if g.exitCode == model.Unset {
		panic("Exit code not set")
	}
	// Set visibility state to false when quitting
	g.setShown(false)
	
	// Safely quit the Fyne app with error recovery
	func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Warn("Recovered from panic during app quit", "panic", r)
			}
		}()
		g.app.Quit()
	}()
	
	_ = removePidFile(g.menuID)
}

// Reset resets the gmenu state without exiting.
// Exiting and restarting is expensive.
func (g *GMenu) Reset(resetInput bool) {
	logrus.Info("resetting gmenu state")

	// Reset the fuse for a new selection cycle - create new fuse first
	newFuse := core.Fuse{}
	g.selectionFuse = newFuse

	// Reset menu state
	if resetInput {
		g.menu.query = ""
		g.ui.SearchEntry.SetText("")
		g.menu.Search("")
	}

	// Reset UI state
	g.ui.SearchEntry.Enable()
	g.exitCode = model.Unset
	g.menu.Selected = 0

	// Safely render UI components with mutex protection
	g.uiMutex.Lock()
	if g.ui != nil && g.ui.ItemsCanvas != nil && g.menu != nil {
		g.ui.ItemsCanvas.Render(g.menu.Filtered, g.menu.Selected, g.config.NoNumericSelection)
		g.ui.MenuLabel.SetText(g.matchCounterLabel())
	}
	g.uiMutex.Unlock()

	logrus.Info("done resetting gmenu state")
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
	defer removePidFile(g.menuID)

	g.app.Run()
	return nil
}

// HideUI hides the UI.
func (g *GMenu) HideUI() {
	if !g.isShown {
		return
	}
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
