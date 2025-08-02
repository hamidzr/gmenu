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

	// Mark selection as made when exit code is set to completion states
	// This ensures WaitForSelection() doesn't hang when SetExitCode is called
	if code != model.Unset {
		g.markSelectionMade()
	}

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

	// Ensure UI is hidden before quitting (in case it wasn't already)
	if g.isShown {
		g.HideUI()
	}

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
	g.resetSelectionFuse()

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
	defer func() {
		if err := removePidFile(g.menuID); err != nil {
			logrus.Errorf("failed to remove PID file: %v", err)
		}
	}()

	g.app.Run()
	return nil
}

// HideUI hides the UI.
func (g *GMenu) HideUI() {
	if !g.isShown {
		return
	}

	// Set flag to prevent OnFocusLost from cancelling during programmatic hide
	g.visibilityMutex.Lock()
	g.isHiding = true
	g.visibilityMutex.Unlock()

	g.ui.MainWindow.Hide()

	// Reset flag and set visibility state
	g.visibilityMutex.Lock()
	g.isHiding = false
	g.isShown = false
	g.visibilityMutex.Unlock()
}

// HideAndReset atomically hides the UI and resets state for reuse
// This ensures consistent behavior between quit and non-quit paths
func (g *GMenu) completeSelection() {
	// Shared logic for completing a selection - used by both quit and non-quit paths
	// This ensures UI is hidden immediately for responsive behavior
	g.HideUI()
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

// resetSelectionFuse creates a new fuse to avoid copy lock issues
func (g *GMenu) resetSelectionFuse() {
	g.selectionFuse = core.Fuse{}
}
