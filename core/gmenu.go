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
	ExitCode      model.ExitCode
	dims          Dimensions
	searchMethod  SearchMethod
	preserveOrder bool
	ui            *GUI
	isRunning     bool
	// SelectionWg is a wait group that lets listeners wait for user being donw with input.
	SelectionWg sync.WaitGroup
	// hasSelection is a flag to prevent multiple Done() calls
	hasSelection bool
	// selectionMutex protects hasSelection
	selectionMutex sync.Mutex
	// isShown tracks whether the UI is currently visible
	isShown bool
	// visibilityMutex protects isShown
	visibilityMutex sync.RWMutex
}

// NewGMenu creates a new GMenu instance.
func NewGMenu(
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
		ExitCode:      model.Unset,
		searchMethod:  searchMethod,
		preserveOrder: conf.PreserveOrder,
		config:        conf,
		store:         store,
		dims: Dimensions{
			MinWidth:  conf.MinWidth,
			MinHeight: conf.MinHeight,
			// max dimensions will be set after UI initialization or from config
			MaxWidth:  conf.MaxWidth,
			MaxHeight: conf.MaxHeight,
		},
		SelectionWg: sync.WaitGroup{},
		isShown:     false, // initially not shown
	}
	g.initUI()
	return g, nil
}

// initValues computes the initial value
func (g *GMenu) initValue(initialQuery string) string {
	lastInput := ""
	if g.menuID != "" && initialQuery == "" {
		cache, err := g.store.LoadCache()
		if err != nil {
			panic(err)
		}
		if canBeHighlighted(cache.LastInput) {
			lastInput = cache.LastInput
		}
	}
	initValue := lastInput
	if initialQuery != "" {
		initValue = initialQuery
	}
	return initValue
}

// SetupMenu sets up the backing menu.
func (g *GMenu) SetupMenu(initialItems []string, initialQuery string) {
	ctx, cancel := context.WithCancel(context.Background())
	submenu, err := newMenu(ctx, initialItems, g.initValue(initialQuery), g.searchMethod, g.preserveOrder)
	if err != nil {
		fmt.Println("Failed to setup menu:", err)
		logrus.Error(err)
		panic(err)
	}
	g.menu = submenu
	g.menuCancel = cancel
	g.setMenuBasedUI()
}

func (g *GMenu) clearCache() error {
	if g.menuID == "" {
		return nil
	}
	cache, err := g.store.LoadCache()
	if err != nil {
		return err
	}
	cache.SetLastInput("")
	cache.SetLastEntry("")
	err = g.store.SaveCache(cache)
	if err != nil {
		return err
	}
	return nil
}

func (g *GMenu) cacheState(value string) error {
	if g.menuID == "" {
		// skip caching if menuID is not set.
		return nil
	}
	cache, err := g.store.LoadCache()
	if err != nil {
		return err
	}
	cache.SetLastInput(g.menu.query)
	cache.SetLastEntry(value)
	err = g.store.SaveCache(cache)
	if err != nil {
		return err
	}
	return nil
}

func (g *GMenu) isUIInitialized() bool {
	return g.ui != nil
}

// one time init for ui elements.
func (g *GMenu) initUI() {
	if g.isUIInitialized() {
		panic("ui is already initialized")
	}
	g.app = app.New()
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
		if g.ExitCode == model.Unset {
			g.ExitCode = model.UserCanceled
			g.markSelectionMade()
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
		if g.ExitCode == model.Unset {
			g.ExitCode = model.UserCanceled
			g.markSelectionMade()
		}
	})

	g.ui = &GUI{
		SearchEntry: searchEntry,
		ItemsCanvas: itemsCanvas,
		MenuLabel:   menuLabel,
		MainWindow:  mainWindow,
	}

}

// markSelectionMade marks that a selection has been made and signals the wait group.
func (g *GMenu) markSelectionMade() {
	g.selectionMutex.Lock()
	defer g.selectionMutex.Unlock()
	if !g.hasSelection {
		g.hasSelection = true
		g.ui.SearchEntry.Disable()
		g.SelectionWg.Done()
	}
}

// ResetUI based on g.menu with minimal rerendering.
func (g *GMenu) setMenuBasedUI() {
	if g.menu == nil || g.ui == nil {
		panic("not initialized")
	}
	g.startListenDynamicUpdates()
	g.ui.SearchEntry.SetText(g.menu.query)
	if g.menu.query != "" {
		g.ui.SearchEntry.SelectAll()
	}
	g.ui.ItemsCanvas.Render(g.menu.Filtered, g.menu.Selected, g.config.NoNumericSelection)
	// show match items out of total item count.
	g.ui.MenuLabel.SetText(g.matchCounterLabel())
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
	// Reset only the state, not the menu
	g.hasSelection = false
	g.SelectionWg = sync.WaitGroup{}
	g.SelectionWg.Add(1)
	g.ExitCode = model.Unset

	// Show window and set focus
	g.ui.MainWindow.Show()
	g.ui.SearchEntry.Enable()
	g.ui.SearchEntry.SetText(g.ui.SearchEntry.Text)

	// Set visibility state
	g.setShown(true)
}
