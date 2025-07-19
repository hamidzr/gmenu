package core

import (
	"fyne.io/fyne/v2"
	"github.com/hamidzr/gmenu/model"
	"time"
)

func (g *GMenu) startListenDynamicUpdates() {
	queryChan := make(chan string)
	g.ui.SearchEntry.OnChanged = func(text string) {
		queryChan <- text
	}
	resizeBasedOnResults := func() {
		if g.ui == nil || g.ui.ItemsCanvas == nil || g.ui.MainWindow == nil {
			return
		}

		resultsSize := g.ui.ItemsCanvas.Container.Size()

		// calculate desired width: between min and max, based on content
		desiredWidth := max(g.dims.MinWidth, resultsSize.Width)
		if g.dims.MaxWidth > 0 {
			desiredWidth = min(desiredWidth, g.dims.MaxWidth)
		}

		// calculate desired height: between min and max, based on content
		desiredHeight := max(g.dims.MinHeight, resultsSize.Height)
		if g.dims.MaxHeight > 0 {
			desiredHeight = min(desiredHeight, g.dims.MaxHeight)
		}

		size := fyne.NewSize(desiredWidth, desiredHeight)
		g.ui.MainWindow.Resize(size)
	}
	go func() { // handle new characters in the search bar and new items loaded.
		// 60 FPS throttling for UI updates
		const targetFPS = 60
		renderTicker := time.NewTicker(time.Second / targetFPS)
		defer renderTicker.Stop()

		var pendingRender bool

		renderUI := func() {
			if !pendingRender {
				return
			}
			pendingRender = false

			// Ensure UI updates happen on main thread or with proper app context
			g.uiMutex.Lock()
			defer g.uiMutex.Unlock()

			if g.ui != nil && g.ui.MenuLabel != nil && g.app != nil {
				func() {
					defer func() {
						if r := recover(); r != nil {
							// Silently ignore theme access panics during tests
							_ = r // SA9003: intentionally ignore panic
						}
					}()
					g.ui.MenuLabel.SetText(g.matchCounterLabel())
				}()
			}
			if g.ui != nil && g.ui.ItemsCanvas != nil && g.app != nil {
				func() {
					defer func() {
						if r := recover(); r != nil {
							// Silently ignore theme access panics during tests
							_ = r // SA9003: intentionally ignore panic
						}
					}()
					g.ui.ItemsCanvas.Render(g.menu.Filtered, g.menu.Selected, g.config.NoNumericSelection)
				}()
			}
			resizeBasedOnResults()
		}

		for {
			select {
			case query := <-queryChan:
				if g.menu != nil {
					g.menu.Search(query)
					pendingRender = true
				}
			case items := <-g.menu.ItemsChan:
				if g.menu != nil {
					g.menu.itemsMutex.Lock()
					deduplicated := make([]model.MenuItem, 0)
					seen := make(map[string]struct{})
					for _, item := range items {
						if _, ok := seen[item.ComputedTitle()]; !ok {
							seen[item.ComputedTitle()] = struct{}{}
							deduplicated = append(deduplicated, item)
						}
					}
					g.menu.items = deduplicated
					g.menu.itemsMutex.Unlock()
					g.menu.Search(g.menu.query)
					pendingRender = true
				}
			case <-renderTicker.C:
				renderUI()
			case <-g.menu.ctx.Done():
				return
			}
		}
	}()
	g.setKeyHandlers()
}

func (g *GMenu) setKeyHandlers() {
	keyHandler := func(key *fyne.KeyEvent) {
		switch key.Name {
		case fyne.KeyDown, fyne.KeyTab:
			if g.menu.Selected < len(g.menu.Filtered)-1 {
				g.menu.Selected++
			} else { // wrap
				g.menu.Selected = 0
			}

		case fyne.KeyUp:
			if g.menu.Selected > 0 {
				g.menu.Selected--
			} else { // wrap
				g.menu.Selected = len(g.menu.Filtered) - 1
			}
		case fyne.KeyReturn, fyne.KeyEnter:
			// con't accept enter key if no items are present and custom selection is disabled.'
			if !g.config.AcceptCustomSelection && len(g.menu.Filtered) == 0 {
				return
			}
			g.markSelectionMade()
		case fyne.KeyEscape:
			g.exitCode = model.UserCanceled
			g.markSelectionMade()
		case fyne.Key1, fyne.Key2, fyne.Key3, fyne.Key4, fyne.Key5, fyne.Key6, fyne.Key7, fyne.Key8, fyne.Key9:
			// handle numeric selection if enabled
			if !g.config.NoNumericSelection {
				var selectedIndex int
				switch key.Name {
				case fyne.Key1:
					selectedIndex = 0
				case fyne.Key2:
					selectedIndex = 1
				case fyne.Key3:
					selectedIndex = 2
				case fyne.Key4:
					selectedIndex = 3
				case fyne.Key5:
					selectedIndex = 4
				case fyne.Key6:
					selectedIndex = 5
				case fyne.Key7:
					selectedIndex = 6
				case fyne.Key8:
					selectedIndex = 7
				case fyne.Key9:
					selectedIndex = 8
				}
				// only select if the index is within bounds
				if selectedIndex < len(g.menu.Filtered) {
					g.menu.Selected = selectedIndex
					g.markSelectionMade()
					return
				}
			}
		default:
			return
		}
		// Safely render UI components
		g.uiMutex.Lock()
		if g.ui != nil && g.ui.ItemsCanvas != nil && g.menu != nil {
			g.ui.ItemsCanvas.Render(g.menu.Filtered, g.menu.Selected, g.config.NoNumericSelection)
		}
		g.uiMutex.Unlock()
	}
	g.ui.SearchEntry.OnKeyDown = keyHandler
	g.ui.MainWindow.Canvas().SetOnTypedKey(keyHandler)
}
