package core

import (
	"context"
	"fmt"
	"sync"

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
	prompt        string
	menuID        string
	menu          *menu
	config        *model.Config
	menuCancel    context.CancelFunc
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
	// isShown tracks whether the UI is currently visible
	isShown bool
	// isHiding tracks when UI is being hidden programmatically to avoid focus loss cancellation
	isHiding        bool
	visibilityMutex sync.RWMutex
}

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
	g.menu = submenu
	g.menuCancel = cancel
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
		g.app = app.New()
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

		if !isHiding {
			// user clicked away or lost focus - cancel the menu and hide immediately
			// IMPORTANT: Set exit code BEFORE marking selection made to avoid race condition
			if g.exitCode == model.Unset {
				g.exitCode = model.UserCanceled
			}
			g.markSelectionMade()
			// Complete selection with shared logic
			g.completeSelection()
		}
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
			g.markSelectionMade()
		}
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
	if g.selectionFuse.Break() {
		// only disable the search entry if we were the one to break the fuse
		g.ui.SearchEntry.Disable()
	}
}

// WaitForSelection waits for the user to make a selection
func (g *GMenu) WaitForSelection() {
	<-g.selectionFuse.Watch()
}

// safeUIUpdate executes a UI update function with proper mutex protection
func (g *GMenu) safeUIUpdate(updateFunc func()) {
	g.uiMutex.Lock()
	defer g.uiMutex.Unlock()
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
	if g.menu == nil || g.ui == nil {
		return fmt.Errorf("menu or UI not initialized")
	}
	g.startListenDynamicUpdates()
	g.safeUIUpdate(func() {
		g.ui.SearchEntry.SetText(g.menu.query)
		if g.menu.query != "" {
			g.ui.SearchEntry.SelectAll()
		}
		g.ui.ItemsCanvas.Render(g.menu.Filtered, g.menu.Selected, g.config.NoNumericSelection)
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
	if g.menu == nil {
		return nil
	}
	g.menu.Search(query)
	return g.menu.Filtered
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
	g.ui.MainWindow.Show()
	g.ui.SearchEntry.Enable()
	g.ui.SearchEntry.SetText(g.ui.SearchEntry.Text)

	// Set focus to the search entry so user can type immediately
	g.ui.MainWindow.Canvas().Focus(g.ui.SearchEntry)

	// Set visibility state
	g.setShown(true)
}
