package core

import (
	"fmt"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hamidzr/gmenu/constant"
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
	app           fyne.App
	store         store.Store
	ExitCode      int
	dims          Dimensions
	mainWindow    fyne.Window
	searchMethod  SearchMethod
	preserveOrder bool
	ui            *GUI
}

// NewGMenu creates a new GMenu instance.
func NewGMenu(
	title string,
	prompt string,
	menu *menu,
	menuID string,
	searchMethod SearchMethod,
	preserveOrder bool,
) (*GMenu, error) {
	store, err := store.NewFileStore[store.Cache, store.Config]([]string{"gmenu", menuID}, "yaml")
	if err != nil {
		return nil, err
	}
	g := &GMenu{
		prompt:        prompt,
		AppTitle:      title,
		menuID:        menuID,
		menu:          menu,
		ExitCode:      constant.UnsetInt,
		searchMethod:  searchMethod,
		preserveOrder: preserveOrder,
		store:         store,
		dims: Dimensions{
			MinWidth:  600,
			MinHeight: 300,
		},
	}
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
	submenu, err := newMenu(initialItems, g.initValue(initialQuery), g.searchMethod, g.preserveOrder)
	if err != nil {
		fmt.Println("Failed to setup menu:", err)
		logrus.Error(err)
		panic(err)
	}
	g.menu = submenu
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

// SetItems sets the items to be displayed in the menu.
func (g *GMenu) SetItems(items []string, serializables []model.GmenuSerializable) {
	menuItems := g.menu.titlesToMenuItem(items)
	for _, item := range serializables {
		myItem := item
		menuItems = append(menuItems, model.MenuItem{AType: &myItem})
	}
	g.menu.itemsMutex.Lock()
	g.menu.ItemsChan <- menuItems
	g.menu.itemsMutex.Unlock()
}

// addItems adds items to the menu.
func (g *GMenu) addItems(items []string, tail bool) {
	newMenuItems := g.menu.titlesToMenuItem(items)
	g.menu.itemsMutex.Lock()
	var newItems []model.MenuItem
	if tail {
		newItems = append(g.menu.items, newMenuItems...)
	} else {
		newItems = append(newMenuItems, g.menu.items...)
	}
	g.menu.itemsMutex.Unlock()
	g.menu.ItemsChan <- newItems
	// TODO: add using SetItems?
}

// PrependItems adds items to the beginning of the menu.
func (g *GMenu) PrependItems(items []string) {
	g.addItems(items, false)
}

// AppendItems adds items to the end of the menu.
func (g *GMenu) AppendItems(items []string) {
	// fmt.Println("appending len items", len(items))
	g.addItems(items, true)
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

// SelectedValue returns the selected item.
func (g *GMenu) SelectedValue() (*model.MenuItem, error) {
	// TODO: check if the app is running. using the doneChan?
	if g.ExitCode == constant.UnsetInt {
		return nil, fmt.Errorf("gmenu has not exited yet")
	}
	// TODO: cli option for allowing query.
	if g.ExitCode != 0 {
		return nil, fmt.Errorf("gmenu exited with code %d", g.ExitCode)
	}
	if g.menu.Selected >= 0 && g.menu.Selected < len(g.menu.Filtered)+1 {
		selected := g.menu.Filtered[g.menu.Selected]
		return &selected, nil
	}
	return &model.MenuItem{Title: g.menu.query}, nil
}

// one time init for ui elements.
func (g *GMenu) initUI() {
	if g.ui != nil {
		panic("ui is already initialized")
	}
	g.app = app.New()
	g.app.Settings().SetTheme(render.MainTheme{Theme: theme.DefaultTheme()})

	g.app.Lifecycle().SetOnExitedForeground(func() {
		if g.ExitCode == constant.UnsetInt {
			g.Quit(1)
		}
	})

	if deskDriver, ok := g.app.Driver().(desktop.Driver); ok {
		g.mainWindow = deskDriver.CreateSplashWindow()
	} else {
		g.mainWindow = g.app.NewWindow(g.AppTitle)
	}
	g.mainWindow.SetTitle(g.AppTitle)
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
	g.mainWindow.SetContent(mainContainer)
	g.mainWindow.Resize(fyne.NewSize(g.dims.MinWidth, g.dims.MinHeight))
	mainContainer.Add(itemsCanvas.Container)
	g.mainWindow.Canvas().Focus(searchEntry)
	g.mainWindow.Show()

	g.ui = &GUI{
		SearchEntry: searchEntry,
		ItemsCanvas: itemsCanvas,
		MenuLabel:   menuLabel,
	}
}

func (g *GMenu) uiBasedOnMenu() {
	queryChan := make(chan string)
	if g.menu == nil || g.ui == nil {
		panic("not initialized")
	}
	g.ui.SearchEntry.OnChanged = func(text string) {
		queryChan <- text
	}
	g.ui.SearchEntry.SetText(g.menu.query)
	if g.menu.query != "" {
		g.ui.SearchEntry.SelectAll()
	}
	g.ui.ItemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
	// show match items out of total item count.
	matchCounterLabel := func() string {
		return fmt.Sprintf("[%d/%d]", g.menu.MatchCount, len(g.menu.items))
	}
	g.ui.MenuLabel.SetText(matchCounterLabel())
	resizeBasedOnResults := func() {
		size := fyne.NewSize(g.dims.MinWidth, g.dims.MinHeight)
		resultsSize := g.ui.ItemsCanvas.Container.Size()
		size.Width = max(g.dims.MinWidth, resultsSize.Width)
		size.Height = resultsSize.Height
		g.mainWindow.Resize(size)
	}
	go func() {
		for {
			select {
			case query := <-queryChan:
				g.menu.Search(query)
				g.ui.MenuLabel.SetText(matchCounterLabel())
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
				g.ui.MenuLabel.SetText(matchCounterLabel())
				g.ui.ItemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
			default:
				// CHECK: should we?
				time.Sleep(10 * time.Millisecond)
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
			g.ui.SearchEntry.Disable()
			g.Quit(0)
		case fyne.KeyEscape:
			g.Quit(1)
		default:
			return
		}
		g.ui.ItemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
	}

	g.ui.SearchEntry.OnKeyDown = keyHandler
	g.mainWindow.Canvas().SetOnTypedKey(keyHandler)
}

// SetupUI creates the UI elements.
func (g *GMenu) SetupUI() {
	g.initUI()
	g.uiBasedOnMenu()
}

// ToggleVisibility toggles the visibility of the gmenu window.
func (g *GMenu) ToggleVisibility() {
	if g.mainWindow.Content().Visible() {
		g.mainWindow.Hide()
	} else {
		g.mainWindow.Show()
	}
}

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
	g.ExitCode = constant.UnsetInt
	g.menu.Selected = 0
	g.SetupMenu([]string{"Loading..."}, "init query")
	g.SetupUI()
}
