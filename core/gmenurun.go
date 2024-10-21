package core

import (
	"fmt"
	"os"

	"github.com/hamidzr/gmenu/constant"
	"github.com/sirupsen/logrus"
)

// Quit exits the application.
func (g *GMenu) Quit(code int) {
	if g.ExitCode != constant.UnsetInt {
		panic("Quit called multiple times")
	}
	g.ExitCode = code
	g.app.Quit()
}

// Reset resets the gmenu state without exiting.
// Exiting and restarting is expensive.
func (g *GMenu) Reset() {
	logrus.Info("resetting gmenu state")
	g.menuCancel()
	g.ExitCode = constant.UnsetInt
	g.menu.Selected = 0
	g.SetupMenu([]string{"Loading..."}, "init query")
}

// Run starts the application.
func (g *GMenu) Run() error {
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
