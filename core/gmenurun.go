package core

import (
	"fmt"
	"os"

	"github.com/hamidzr/gmenu/constant"
	"github.com/sirupsen/logrus"
)


// MakeSelection makes a selection.
func (g *GMenu) MakeSelection(idx int) {
	// g.ResultChan <- RunResult{ExitCode: 0}
	g.Quit(0)
}

// Quit exits the application.
func (g *GMenu) Quit(code int) {
	// if g.ExitCode != constant.UnsetInt {
	// 	panic("Quit called multiple times")
	// }
	g.ResultChan <- RunResult{ExitCode: code}
	g.ExitCode = code
	g.app.Quit()
}

// Reset resets the gmenu state without exiting.
// Exiting and restarting is expensive.
func (g *GMenu) Reset() {
	logrus.Info("resetting gmenu state")
	g.menuCancel()
	g.ExitCode = constant.UnsetInt
	g.ResultChan = make(chan RunResult, 1)
	g.menu.Selected = 0
	g.SetupMenu([]string{"Loading..."}, "init_query")
}

// Run starts the application.
func (g *GMenu) Run() error {
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
