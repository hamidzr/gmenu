package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/frostbyte73/core"
	"github.com/hamidzr/gmenu/model"
	"github.com/hamidzr/gmenu/render"
	"github.com/hamidzr/gmenu/store"
	"github.com/sirupsen/logrus"
)

// Dimensions define geometry of the application window.
type Dimensions struct {
	MinWidth  float32
	MinHeight float32
	MaxWidth  float32
	MaxHeight float32
}

// GUI aka GMenuUI holds ui pieces.
type GUI struct {
	MainWindow  fyne.Window
	SearchEntry *render.SearchEntry
	ItemsCanvas *render.ItemsCanvas
	MenuLabel   *widget.Label
}

// GMenu is the main application struct for GoMenu.
type GMenu struct {
	AppTitle string
	// prompt that shows up in the search bar.
	prompt     string
	menuID     string
	menu       *menu
	config     *model.Config
	menuCancel context.CancelFunc
	// menuMutex protects access to menu and menuCancel swapping
	menuMutex     sync.RWMutex
	app           fyne.App
	store         store.Store
	exitCode      model.ExitCode
	dims          Dimensions
	searchMethod  SearchMethod
	preserveOrder bool
	ui            *GUI
	uiMutex       sync.Mutex
	isRunning     bool
	// selectionFuse is a one-way switch that can only be broken once
	selectionFuse core.Fuse
	// selectionMutex guards selectionFuse operations and resets
	selectionMutex sync.Mutex
	// isShown tracks whether the UI is currently visible
	isShown bool
	// isHiding tracks when UI is being hidden programmatically to avoid focus loss cancellation
	isHiding        bool
	visibilityMutex sync.RWMutex
}

// newAppFunc creates a new fyne App. Overridden in tests to use fyne test app.
var newAppFunc = func() fyne.App { return app.New() }

// Note: UI serialization is handled via render.UIRenderMutex to ensure
// all UI interactions (including those originating from tests) share the
// same critical section.

// NewGMenu creates a new GMenu instance.
func NewGMenu(
	searchMethod SearchMethod,
	conf *model.Config,
) (*GMenu, error) {
	return NewGMenuWithApp(nil, searchMethod, conf)
}

// NewGMenuWithApp creates a new GMenu instance with a specific Fyne app (useful for testing).
func NewGMenuWithApp(
	fyneApp fyne.App,
	searchMethod SearchMethod,
	conf *model.Config,
) (*GMenu, error) {
	store, err := store.NewFileStore[store.Cache, store.Config]([]string{"gmenu", conf.MenuID}, "yaml")
	if err != nil {
		return nil, err
	}
	g := &GMenu{
		prompt:        conf.Prompt,
		AppTitle:      conf.Title,
		menuID:        conf.MenuID,
		exitCode:      model.Unset,
		searchMethod:  searchMethod,
		preserveOrder: conf.PreserveOrder,
		config:        conf,
		store:         store,
		app:           fyneApp, // Use provided app if available
		dims: Dimensions{
			MinWidth:  conf.MinWidth,
			MinHeight: conf.MinHeight,
			// max dimensions will be set after UI initialization or from config
			MaxWidth:  conf.MaxWidth,
			MaxHeight: conf.MaxHeight,
		},
		// selectionFuse is initialized as zero value (ready to be broken)
		isShown: false, // initially not shown
	}
	if err := g.initUI(); err != nil {
		return nil, fmt.Errorf("failed to initialize UI: %w", err)
	}
	return g, nil
}

func (g *GMenu) GetExitCode() model.ExitCode {
	return g.exitCode
}

// initValue computes the initial value for the search query
func (g *GMenu) initValue(initialQuery string) (string, error) {
	lastInput := ""
	if g.menuID != "" && initialQuery == "" {
		cache, err := g.store.LoadCache()
		if err != nil {
			logrus.Warn("Failed to load cache for initial value:", err)
			// continue with empty lastInput rather than failing
		} else if canBeHighlighted(cache.LastInput) {
			lastInput = cache.LastInput
		}
	}
	initValue := lastInput
	if initialQuery != "" {
		initValue = initialQuery
	}
	return initValue, nil
}

// SetupMenu sets up the backing menu.
func (g *GMenu) SetupMenu(initialItems []string, initialQuery string) error {
	ctx, cancel := context.WithCancel(context.Background())
	initVal, err := g.initValue(initialQuery)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to get initial value: %w", err)
	}
	submenu, err := newMenu(ctx, initialItems, initVal, g.searchMethod, g.preserveOrder)
	if err != nil {
		cancel()
		logrus.Error("Failed to setup menu:", err)
		return fmt.Errorf("failed to create menu: %w", err)
	}
	// Cancel existing and swap under lock
	g.menuMutex.Lock()
	if g.menuCancel != nil {
		g.menuCancel()
	}
	g.menu = submenu
	g.menuCancel = cancel
	g.menuMutex.Unlock()
	if err := g.setMenuBasedUI(); err != nil {
		cancel()
		return fmt.Errorf("failed to setup UI: %w", err)
	}
	return nil
}

func (g *GMenu) clearCache() error {
	return g.withCache(func(cache *store.Cache) error {
		cache.SetLastInput("")
		cache.SetLastEntry("")
		return nil
	})
}

// ensureSelectionExitCode updates the exit code for a completed selection.
// It preserves explicit cancellations but upgrades pending/optimistic states to success.
func (g *GMenu) ensureSelectionExitCode(code model.ExitCode) {
	g.selectionMutex.Lock()
	switch g.exitCode {
	case model.Unset:
		g.exitCode = code
	case model.UserCanceled:
		if code == model.NoError {
			g.exitCode = code
		}
	}
	g.selectionMutex.Unlock()
}

func (g *GMenu) cacheState(value string) error {
	return g.withCache(func(cache *store.Cache) error {
		cache.SetLastInput(g.menu.query)
		cache.SetLastEntry(value)
		return nil
	})
}

func (g *GMenu) isUIInitialized() bool {
	return g.ui != nil
}

// initUI initializes UI elements - should only be called once
func (g *GMenu) initUI() error {
	if g.isUIInitialized() {
		return fmt.Errorf("ui is already initialized")
	}
	if g.app == nil {
		g.app = newAppFunc()
	}
	g.app.Settings().SetTheme(render.MainTheme{Theme: theme.DefaultTheme()})

	// g.app.Lifecycle().SetOnExitedForeground(func() {
	// 	if g.ExitCode == constant.UnsetInt {
	// 		g.Quit(1)
	// 	}
	// })
	var mainWindow fyne.Window

	if deskDriver, ok := g.app.Driver().(desktop.Driver); ok {
		mainWindow = deskDriver.CreateSplashWindow()
	} else {
		mainWindow = g.app.NewWindow(g.AppTitle)
	}
	mainWindow.SetTitle(g.AppTitle)
	entryDisabledKeys := map[fyne.KeyName]bool{
		fyne.KeyUp:   true,
		fyne.KeyDown: true,
		fyne.KeyTab:  true,
	}
	searchEntry := &render.SearchEntry{PropagationBlacklist: entryDisabledKeys}
	searchEntry.ExtendBaseWidget(searchEntry)
	searchEntry.SetPlaceHolder(g.prompt)
	searchEntry.OnFocusLost = func() {
		// only cancel on focus loss if not being hidden programmatically
		g.visibilityMutex.RLock()
		isHiding := g.isHiding
		g.visibilityMutex.RUnlock()

		if isHiding {
			return
		}

		go func() {
			// Give legitimate selections a chance to finish before treating focus loss as cancel
			const focusLossGrace = 40 * time.Millisecond
			timer := time.NewTimer(focusLossGrace)
			defer timer.Stop()
			<-timer.C

			if g.selectionFuse.IsBroken() {
				return
			}

			g.selectionMutex.Lock()
			if g.exitCode == model.Unset {
				g.exitCode = model.UserCanceled
			}
			g.selectionMutex.Unlock()

			g.markSelectionMade()
			// Complete selection with shared logic
			g.completeSelection()
		}()
	}
	itemsCanvas := render.NewItemsCanvas()
	menuLabel := widget.NewLabel("menulabel")
	inputBox := render.NewInputArea(searchEntry, menuLabel)
	mainContainer := container.NewVBox(inputBox)
	mainWindow.SetContent(mainContainer)
	mainWindow.Resize(fyne.NewSize(g.dims.MinWidth, g.dims.MinHeight))
	mainContainer.Add(itemsCanvas.Container)
	mainWindow.Canvas().Focus(searchEntry)

	// Add focus loss detection using OnClose
	mainWindow.SetOnClosed(func() {
		if g.exitCode == model.Unset {
			g.exitCode = model.UserCanceled
		}
		// Always call markSelectionMade to ensure the fuse is broken
		g.markSelectionMade()
	})

	g.ui = &GUI{
		SearchEntry: searchEntry,
		ItemsCanvas: itemsCanvas,
		MenuLabel:   menuLabel,
		MainWindow:  mainWindow,
	}

	return nil
}

// markSelectionMade marks that a selection has been made by breaking the fuse.
func (g *GMenu) markSelectionMade() {
	// break the fuse - this can only happen once and is thread-safe
	g.selectionMutex.Lock()
	broke := g.selectionFuse.Break()
	g.selectionMutex.Unlock()
	if broke {
		// only disable the search entry if we were the one to break the fuse
		g.safeUIUpdate(func() {
			if g.ui != nil && g.ui.SearchEntry != nil {
				g.ui.SearchEntry.Disable()
			}
		})
	}
}

// handleItemClick handles when a user clicks on a menu item
func (g *GMenu) handleItemClick(index int) {
	// Protect navigation state with menu items mutex
	g.menu.itemsMutex.Lock()
	if index >= 0 && index < len(g.menu.Filtered) {
		g.menu.Selected = index
	}
	g.menu.itemsMutex.Unlock()

	g.ensureSelectionExitCode(model.NoError)
	// Complete the selection like keyboard Enter
	g.markSelectionMade()
	g.completeSelection()
}

// WaitForSelection waits for the user to make a selection
func (g *GMenu) WaitForSelection() {
	<-g.selectionFuse.Watch()
}

// safeUIUpdate executes a UI update function with proper mutex protection
func (g *GMenu) safeUIUpdate(updateFunc func()) {
	// Serialize per-instance and marshal onto main thread when possible
	g.uiMutex.Lock()
	defer g.uiMutex.Unlock()
	if g.app != nil && g.app.Driver() != nil {
		if runner, ok := g.app.Driver().(interface{ RunOnMain(func()) }); ok {
			done := make(chan struct{})
			runner.RunOnMain(func() {
				updateFunc()
				close(done)
			})
			<-done
			return
		}
	}
	updateFunc()
}

// withCache executes an operation on the cache and saves it back
func (g *GMenu) withCache(operation func(*store.Cache) error) error {
	if g.menuID == "" {
		return nil // skip caching if menuID is not set
	}
	cache, err := g.store.LoadCache()
	if err != nil {
		return err
	}

	if err := operation(&cache); err != nil {
		return err
	}

	return g.store.SaveCache(cache)
}

// setMenuBasedUI updates UI based on g.menu with minimal rerendering.
func (g *GMenu) setMenuBasedUI() error {
	g.menuMutex.RLock()
	currentMenu := g.menu
	g.menuMutex.RUnlock()
	if currentMenu == nil || g.ui == nil {
		return fmt.Errorf("menu or UI not initialized")
	}
	// Start listeners bound to the current menu snapshot so later swaps don't race
	g.startListenDynamicUpdatesForMenu(currentMenu)
	g.safeUIUpdate(func() {
		// Read query under its lock to avoid data race
		currentMenu.queryMutex.Lock()
		currentQuery := currentMenu.query
		currentMenu.queryMutex.Unlock()
		g.ui.SearchEntry.SetText(currentQuery)
		if currentQuery != "" {
			g.ui.SearchEntry.SelectAll()
		}
		// Read filtered state under items lock for consistency
		currentMenu.itemsMutex.Lock()
		filtered := append([]model.MenuItem(nil), currentMenu.Filtered...)
		selected := currentMenu.Selected
		currentMenu.itemsMutex.Unlock()
		g.ui.ItemsCanvas.Render(filtered, selected, g.config.NoNumericSelection, g.handleItemClick)
		// show match items out of total item count.
		g.ui.MenuLabel.SetText(g.matchCounterLabel())
	})
	return nil
}

// ToggleVisibility toggles the visibility of the gmenu window.
func (g *GMenu) ToggleVisibility() {
	if g.IsShown() {
		g.HideUI()
	} else {
		g.ShowUI()
	}
}

// Search performs a search on the menu items using the configured search method.
func (g *GMenu) Search(query string) []model.MenuItem {
	g.menuMutex.RLock()
	m := g.menu
	g.menuMutex.RUnlock()
	if m == nil {
		return []model.MenuItem{}
	}
	m.Search(query)
	// snapshot results under items lock for thread-safety
	m.itemsMutex.Lock()
	filtered := append([]model.MenuItem{}, m.Filtered...)
	m.itemsMutex.Unlock()
	return filtered
}

// IsShown returns whether the UI is currently visible
func (g *GMenu) IsShown() bool {
	g.visibilityMutex.RLock()
	defer g.visibilityMutex.RUnlock()
	return g.isShown
}

// setShown sets the visibility state with proper locking
func (g *GMenu) setShown(shown bool) {
	g.visibilityMutex.Lock()
	defer g.visibilityMutex.Unlock()
	g.isShown = shown
}

// ShowUI and wait for user input.
func (g *GMenu) ShowUI() {

	// Show window and set focus
	g.safeUIUpdate(func() {
		if g.ui == nil || g.ui.MainWindow == nil || g.ui.SearchEntry == nil {
			return
		}
		g.ui.MainWindow.Show()
		g.ui.SearchEntry.Enable()
		g.ui.SearchEntry.SetText(g.ui.SearchEntry.Text)
	})

	// Set focus to the search entry so user can type immediately
	g.safeUIUpdate(func() {
		if g.ui != nil && g.ui.MainWindow != nil && g.ui.SearchEntry != nil {
			g.ui.MainWindow.Canvas().Focus(g.ui.SearchEntry)
		}
	})

	// Set visibility state
	g.setShown(true)
}
