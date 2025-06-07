package core

import (
	"fyne.io/fyne/v2"
	"github.com/hamidzr/gmenu/model"
)

func (g *GMenu) startListenDynamicUpdates() {
	queryChan := make(chan string)
	g.ui.SearchEntry.OnChanged = func(text string) {
		queryChan <- text
	}
	resizeBasedOnResults := func() {
		size := fyne.NewSize(g.dims.MinWidth, g.dims.MinHeight)
		resultsSize := g.ui.ItemsCanvas.Container.Size()
		size.Width = max(g.dims.MinWidth, resultsSize.Width)
		size.Height = resultsSize.Height
		g.ui.MainWindow.Resize(size)
	}
	go func() { // handle new characters in the search bar and new items loaded.
		for {
			select {
			case query := <-queryChan:
				g.menu.Search(query)
				g.ui.MenuLabel.SetText(g.matchCounterLabel())
				g.ui.ItemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
				resizeBasedOnResults()
			case items := <-g.menu.ItemsChan:
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
				g.ui.MenuLabel.SetText(g.matchCounterLabel())
				g.ui.ItemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
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
		case fyne.KeyReturn:
			// con't accept enter key if no items are present and custom selection is disabled.'
			if !g.config.AcceptCustomSelection && len(g.menu.Filtered) == 0 {
				return
			}
			g.markSelectionMade()
		case fyne.KeyEscape:
			g.ExitCode = model.UserCanceled
			g.markSelectionMade()
		default:
			return
		}
		g.ui.ItemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
	}
	g.ui.SearchEntry.OnKeyDown = keyHandler
	g.ui.MainWindow.Canvas().SetOnTypedKey(keyHandler)
}
