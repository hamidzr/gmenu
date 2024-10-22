package core

import (
	"fmt"
	"sync"

	"github.com/hamidzr/gmenu/constant"
	"github.com/sirupsen/logrus"
)

// setExitCode sets the exit code for the application.
func (g *GMenu) SetExitCode(code int) {
	logrus.Debug("setting exit code to: ", code)
	if g.ExitCode != constant.UnsetInt && g.ExitCode != code {
		panic("Exit code set multiple times to different values")
	}
	g.ExitCode = code
}

// Quit exits the application.
func (g *GMenu) Quit(code int) {
	g.SetExitCode(code)
	g.app.Quit()
	_ = removePidFile(g.menuID)
}

// Reset resets the gmenu state without exiting.
// Exiting and restarting is expensive.
func (g *GMenu) Reset() {
	logrus.Info("resetting gmenu state")
	g.menuCancel()
	g.ui.SearchEntry.Enable()
	g.ExitCode = constant.UnsetInt
	g.SelectionWg = sync.WaitGroup{}
	g.menu.Selected = 0
	removePidFile(g.menuID)
	g.SetupMenu([]string{"Loading..."}, "")
}

func (g *GMenu) RunAppForever() error {
	if g.isRunning {
		panic("Run called multiple times")
	}
	g.isRunning = true
	g.app.Run()
	return nil
}

// ShowUI and wait for user input.
func (g *GMenu) ShowUI() {
	g.SelectionWg.Add(1)
	g.mainWindow.Show()
	_, err := createPidFile(g.menuID)
	if err != nil {
		g.Quit(1)
	}
}

// HideUI hides the UI.
func (g *GMenu) HideUI() {
	g.mainWindow.Hide()
	removePidFile(g.menuID)
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
