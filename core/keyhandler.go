package core

import (
	"time"

	"fyne.io/fyne/v2"
	"github.com/hamidzr/gmenu/model"
)

const (
	queryChannelBufferSize = 100 // buffer size for query updates channel
)

// numericKeyToIndex converts numeric key names to zero-based indices
func numericKeyToIndex(keyName fyne.KeyName) (int, bool) {
	switch keyName {
	case fyne.Key1:
		return 0, true
	case fyne.Key2:
		return 1, true
	case fyne.Key3:
		return 2, true
	case fyne.Key4:
		return 3, true
	case fyne.Key5:
		return 4, true
	case fyne.Key6:
		return 5, true
	case fyne.Key7:
		return 6, true
	case fyne.Key8:
		return 7, true
	case fyne.Key9:
		return 8, true
	default:
		return 0, false
	}
}

// startListenDynamicUpdatesForMenu wires listeners for a specific menu instance.
// Passing the menu explicitly avoids races when g.menu is swapped concurrently.
func (g *GMenu) startListenDynamicUpdatesForMenu(m *menu) {
	queryChan := make(chan string, queryChannelBufferSize) // buffered channel to prevent blocking
	// Assign UI handler under UI mutex to avoid races when multiple setups occur
	g.uiMutex.Lock()
	g.ui.SearchEntry.OnChanged = func(text string) {
		select {
		case queryChan <- text:
		default:
			// drop update if channel is full to prevent blocking
		}
	}
	g.uiMutex.Unlock()
	// Dynamic resize disabled in tests to reduce UI races
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
			g.safeUIUpdate(func() {
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
						// snapshot filtered data under menu lock for consistent render
						m.itemsMutex.Lock()
						filtered := append([]model.MenuItem(nil), m.Filtered...)
						selected := m.Selected
						m.itemsMutex.Unlock()
						g.ui.ItemsCanvas.Render(filtered, selected, g.config.NoNumericSelection, g.handleItemClick)
					}()
				}
				// Disabled dynamic resizing during tests to avoid UI races
			})
		}

		for {
			select {
			case query := <-queryChan:
				if m != nil {
					m.Search(query)
					pendingRender = true
				}
			case items := <-m.ItemsChan:
				if m != nil {
					// Deduplicate and replace items under items lock
					m.itemsMutex.Lock()
					deduplicated := make([]model.MenuItem, 0, len(items))
					seen := make(map[string]struct{}, len(items))
					for _, item := range items {
						key := item.ComputedTitle()
						if _, ok := seen[key]; !ok {
							seen[key] = struct{}{}
							deduplicated = append(deduplicated, item)
						}
					}
					m.items = deduplicated
					// capture current query while holding query lock to search consistently
					m.queryMutex.Lock()
					currentQuery := m.query
					m.queryMutex.Unlock()
					m.itemsMutex.Unlock()

					// Re-run search with current query (Search handles locking)
					m.Search(currentQuery)
					pendingRender = true
				}
			case <-renderTicker.C:
				renderUI()
			case <-m.ctx.Done():
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
			// Protect navigation state with menu items mutex
			g.menu.itemsMutex.Lock()
			if g.menu.Selected < len(g.menu.Filtered)-1 {
				g.menu.Selected++
			} else { // wrap
				g.menu.Selected = 0
			}
			g.menu.itemsMutex.Unlock()

		case fyne.KeyUp:
			// Protect navigation state with menu items mutex
			g.menu.itemsMutex.Lock()
			if g.menu.Selected > 0 {
				g.menu.Selected--
			} else { // wrap
				g.menu.Selected = len(g.menu.Filtered) - 1
			}
			g.menu.itemsMutex.Unlock()
		case fyne.KeyReturn, fyne.KeyEnter:
			// con't accept enter key if no items are present and custom selection is disabled.'
			if !g.config.AcceptCustomSelection && len(g.menu.Filtered) == 0 {
				return
			}
			g.ensureSelectionExitCode(model.NoError)
			g.markSelectionMade()
			// Complete selection with shared logic
			g.completeSelection()
		case fyne.KeyEscape:
			g.exitCode = model.UserCanceled
			g.markSelectionMade()
			// Complete selection with shared logic
			g.completeSelection()
		case fyne.Key1, fyne.Key2, fyne.Key3, fyne.Key4, fyne.Key5, fyne.Key6, fyne.Key7, fyne.Key8, fyne.Key9:
			// handle numeric selection if enabled
			if !g.config.NoNumericSelection {
				if selectedIndex, ok := numericKeyToIndex(key.Name); ok {
					// only select if the index is within bounds
					if selectedIndex < len(g.menu.Filtered) {
						g.menu.Selected = selectedIndex
						g.ensureSelectionExitCode(model.NoError)
						g.markSelectionMade()
						// Complete selection with shared logic
						g.completeSelection()
						return
					}
				}
			}
		default:
			return
		}
		// Safely render UI components
		g.uiMutex.Lock()
		if g.ui != nil && g.ui.ItemsCanvas != nil && g.menu != nil {
			// snapshot filtered data under lock for consistent render
			g.menu.itemsMutex.Lock()
			filtered := append([]model.MenuItem(nil), g.menu.Filtered...)
			selected := g.menu.Selected
			g.menu.itemsMutex.Unlock()
			g.ui.ItemsCanvas.Render(filtered, selected, g.config.NoNumericSelection, g.handleItemClick)
		}
		g.uiMutex.Unlock()
	}
	// Assign under UI mutex to avoid concurrent writes in tests
	g.uiMutex.Lock()
	g.ui.SearchEntry.OnKeyDown = keyHandler
	g.uiMutex.Unlock()
	// Note: MainWindow.Canvas().SetOnTypedKey() removed to prevent double key processing
	// SearchEntry handles all keys via OnKeyDown and PropagationBlacklist
}
