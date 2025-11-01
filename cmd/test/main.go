package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	splashDuration = 1500 * time.Millisecond
	mainWidth      = 420
	mainHeight     = 160
)

func main() {
	fyneApp := app.NewWithID("gmenu-idle-wakeup-test")

	mainWindow := fyneApp.NewWindow("Idle Wakeup Probe")
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Type here to watch idle wakeups in your profiler")
	mainWindow.SetContent(container.NewVBox(
		widget.NewLabel("Type to generate UI events and monitor idle wakeups"),
		entry,
	))
	mainWindow.Resize(fyne.NewSize(mainWidth, mainHeight))
	mainWindow.SetOnClosed(func() {
		fyneApp.Quit()
	})

	splash := fyneApp.NewWindow("Starting gmenu test")
	splash.SetContent(container.NewCenter(canvas.NewText(
		"Preparing probe...",
		theme.PrimaryColor(),
	)))
	splash.SetFixedSize(true)

	go func() {
		time.Sleep(splashDuration)
		if current := fyne.CurrentApp(); current != nil {
			if driver := current.Driver(); driver != nil {
				if runner, ok := driver.(interface{ RunOnMain(func()) }); ok {
					runner.RunOnMain(func() {
						mainWindow.Show()
						splash.Close()
					})
					return
				}
			}
		}
		mainWindow.Show()
		splash.Close()
	}()

	splash.Show()
	fyneApp.Run()
}
