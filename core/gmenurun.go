package core

import (
	"fmt"
	"os"
	"sync"

	"github.com/hamidzr/gmenu/constant"
	"github.com/sirupsen/logrus"
)

// // HideOnSelection hides the application when a selection is made.
// func (g *GMenu) HideAndResetOnSelection() {
// 	g.SelectionWg.Wait()
// 	g.mainWindow.Hide()
// }
//

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
	g.SetupMenu([]string{"Loading..."}, "init_query")
}

func (g *GMenu) RunAppForever() error {
	if g.isRunning {
		panic("Run called multiple times")
	}
	g.isRunning = true
	pidFile, err := createPidFile(g.menuID)
	defer func() { // clean up the pid file.
		if pidFile != "" {
			if err := os.Remove(pidFile); err != nil {
				fmt.Println("Failed to remove pid file:", pidFile)
				logrus.Error(err)
			}
		}
	}()
	if err != nil {
		g.Quit(1)
		return err
	}
	g.app.Run()
	return nil
}

func (g *GMenu) ShowUI() {
	g.SelectionWg.Add(1)
	g.mainWindow.Show()
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
