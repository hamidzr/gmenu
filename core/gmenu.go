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
	config        model.Config
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
}

// NewGMenu creates a new GMenu instance.
func NewGMenu(
	searchMethod SearchMethod,
	conf model.Config,
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
			MinWidth:  600,
			MinHeight: 300,
		},
		SelectionWg: sync.WaitGroup{},
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
	itemsCanvas := render.NewItemsCanvas()
	menuLabel := widget.NewLabel("menulabel")
	inputBox := render.NewInputArea(searchEntry, menuLabel)
	mainContainer := container.NewVBox(inputBox)
	mainWindow.SetContent(mainContainer)
	mainWindow.Resize(fyne.NewSize(g.dims.MinWidth, g.dims.MinHeight))
	mainContainer.Add(itemsCanvas.Container)
	mainWindow.Canvas().Focus(searchEntry)

	g.ui = &GUI{
		SearchEntry: searchEntry,
		ItemsCanvas: itemsCanvas,
		MenuLabel:   menuLabel,
		MainWindow:  mainWindow,
	}
}

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
			g.ui.SearchEntry.Disable()
			g.SelectionWg.Done()
		case fyne.KeyEscape:
			g.ExitCode = model.UserCanceled
			g.SelectionWg.Done()
		default:
			return
		}
		g.ui.ItemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
	}
	g.ui.SearchEntry.OnKeyDown = keyHandler
	g.ui.MainWindow.Canvas().SetOnTypedKey(keyHandler)
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
	g.ui.ItemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
	// show match items out of total item count.
	g.ui.MenuLabel.SetText(g.matchCounterLabel())
}

// ToggleVisibility toggles the visibility of the gmenu window.
func (g *GMenu) ToggleVisibility() {
	if g.ui.MainWindow.Content().Visible() {
		g.HideUI()
	} else {
		g.ShowUI()
	}
}
