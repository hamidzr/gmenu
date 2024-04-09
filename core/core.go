package core

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/hamidzr/gmenu/model"
	"github.com/hamidzr/gmenu/render"
	"github.com/hamidzr/gmenu/store"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

/*
TODO:
- on failure clean up the pid file

*/

const (
	unsetInt = -1
)

type menu struct {
	items      []model.MenuItem
	query      string
	itemsMutex sync.Mutex
	queryMutex sync.Mutex
	ItemsChan  chan []model.MenuItem

	Filtered []model.MenuItem
	// zero-based index of the selected item in the filtered list
	Selected int
	// ResultText   string
	// MatchCount is the number of items that matched the search query.
	MatchCount    int
	SearchMethod  SearchMethod
	resultLimit   int
	preserveOrder bool
}

func newMenu(
	itemTitles []string,
	initValue string,
	searchMethod SearchMethod,
	preserveOrder bool,
) *menu {
	m := menu{
		Selected:      0,
		SearchMethod:  searchMethod,
		resultLimit:   10,
		ItemsChan:     make(chan []model.MenuItem),
		query:         initValue,
		preserveOrder: preserveOrder,
	}
	items := m.titlesToMenuItem(itemTitles)
	m.items = items

	if len(items) == 0 {
		panic("Menu must have at least one item")
	}

	m.Search(initValue)
	return &m
}

// Filters the menu filtered list to only include items that match the keyword.
func (m *menu) Search(keyword string) {
	m.queryMutex.Lock()
	m.query = keyword
	m.queryMutex.Unlock()
	if keyword == "" {
		m.Filtered = m.items
	} else {
		// start := time.Now()
		m.Filtered = m.SearchMethod(m.items, keyword, m.preserveOrder, m.resultLimit)
		// elapsed := time.Since(start)
		// fmt.Println("Search took", elapsed)
	}
	if len(m.Filtered) > 0 {
		m.Selected = 0
	} else {
		m.Selected = unsetInt
	}
	m.MatchCount = len(m.Filtered)
	if len(m.Filtered) > m.resultLimit {
		m.Filtered = m.Filtered[:m.resultLimit]
	}
}

func (m *menu) titlesToMenuItem(titles []string) []model.MenuItem {
	items := make([]model.MenuItem, len(titles))
	for i, entry := range titles {
		items[i] = model.MenuItem{Title: entry}
	}
	return items
}

// Dimensions define geometry of the application window.
type Dimensions struct {
	MinWidth  float32
	MinHeight float32
}

// GMenu is the main application struct for GoMenu.
type GMenu struct {
	AppTitle string
	prompt   string
	menuID   string
	menu     *menu
	app      fyne.App
	store    store.Store
	ExitCode int
	dims     Dimensions
}

// NewGMenu creates a new GMenu instance.
func NewGMenu(
	initialItems []string,
	title string,
	prompt string,
	menuID string,
	searchMethod SearchMethod,
	preserveOrder bool,
) (*GMenu, error) {
	store, err := store.NewFileStore[store.Cache, store.Config]([]string{"gmenu", menuID})
	if err != nil {
		return nil, err
	}
	lastInput := ""
	if menuID != "" {
		cache, err := store.LoadCache()
		if err != nil {
			return nil, err
		}
		if canBeHighlighted(cache.LastInput) {
			lastInput = cache.LastInput
		}

	}
	menu := newMenu(initialItems, lastInput, searchMethod, preserveOrder)
	g := &GMenu{
		prompt:   prompt,
		AppTitle: title,
		menuID:   menuID,
		ExitCode: unsetInt,
		menu:     menu,
		store:    store,
		dims: Dimensions{
			MinWidth:  600,
			MinHeight: 300,
		},
	}
	g.setupUI()
	return g, nil
}

// canBeHighlighted returns true if the menu item can be highlighted
// programmatically via exiting fayne interface.
func canBeHighlighted(entry string) bool {
	// TODO: find a better way to select all on searchEnty.
	for _, c := range entry {
		if !(c >= 'a' && c <= 'z' ||
			c >= 'A' && c <= 'Z' ||
			c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

// Run starts the application.
func (g *GMenu) Run() error {
	pidFile, err := createPidFile(g.menuID)
	if err != nil {
		g.Quit(1)
		return err
	}
	g.app.Run()
	if pidFile != "" {
		if err := os.Remove(pidFile); err != nil {
			fmt.Println("Failed to remove pid file:", pidFile)
			return err
		}
	}
	selectedVal, err := g.SelectedValue()
	if err != nil {
		if cacheErr := g.clearCache(); cacheErr != nil {
			fmt.Println("Failed to clear cache:", cacheErr)
		}
		return err
	}
	err = g.cacheState(selectedVal)
	return err
}

// SetItems sets the items to be displayed in the menu.
func (g *GMenu) SetItems(items []string) {
	g.menu.itemsMutex.Lock()
	g.menu.ItemsChan <- g.menu.titlesToMenuItem(items)
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
func (g *GMenu) SelectedValue() (string, error) {
	// TODO: check if the app is running. using the doneChan?
	if g.ExitCode == unsetInt {
		return "", fmt.Errorf("gmenu has not exited yet")
	}
	// TODO: cli option for allowing query.
	if g.ExitCode != 0 {
		return "", fmt.Errorf("gmenu exited with code %d", g.ExitCode)
	}
	if g.menu.Selected >= 0 && g.menu.Selected < len(g.menu.Filtered)+1 {
		selected := g.menu.Filtered[g.menu.Selected].Title
		return selected, nil
	}
	return g.menu.query, nil
}

// Quit exits the application.
func (g *GMenu) Quit(code int) {
	g.ExitCode = code
	g.app.Quit()
}

// setupUI creates the UI elements.
func (g *GMenu) setupUI() {
	queryChan := make(chan string)

	g.app = app.New()
	g.app.Settings().SetTheme(render.MainTheme{Theme: theme.DefaultTheme()})

	g.app.Lifecycle().SetOnExitedForeground(func() {
		if g.ExitCode == unsetInt {
			g.Quit(1)
		}
	})

	var myWindow fyne.Window
	if deskDriver, ok := g.app.Driver().(desktop.Driver); ok {
		myWindow = deskDriver.CreateSplashWindow()
	} else {
		myWindow = g.app.NewWindow(g.AppTitle)
	}
	myWindow.SetTitle(g.AppTitle)
	entryDisabledKeys := map[fyne.KeyName]bool{
		fyne.KeyUp:   true,
		fyne.KeyDown: true,
		fyne.KeyTab:  true,
	}
	searchEntry := &render.SearchEntry{PropagationBlacklist: entryDisabledKeys}
	searchEntry.ExtendBaseWidget(searchEntry)
	searchEntry.SetPlaceHolder(g.prompt)
	searchEntry.SetText(g.menu.query)
	if g.menu.query != "" {
		searchEntry.SelectAll()
	}
	itemsCanvas := render.NewItemsCanvas()
	itemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
	// show match items out of total item count.
	matchCounterLabel := func() string {
		return fmt.Sprintf("[%d/%d]", g.menu.MatchCount, len(g.menu.items))
	}
	menuLabel := widget.NewLabel(matchCounterLabel())

	inputBox := render.NewInputArea(searchEntry, menuLabel)
	mainContainer := container.NewVBox(inputBox)
	myWindow.SetContent(mainContainer)

	searchEntry.OnChanged = func(text string) {
		queryChan <- text
	}

	resizeBasedOnResults := func() {
		size := fyne.NewSize(g.dims.MinWidth, g.dims.MinHeight)
		resultsSize := itemsCanvas.Container.Size()
		size.Width = max(g.dims.MinWidth, resultsSize.Width)
		size.Height = resultsSize.Height
		myWindow.Resize(size)
	}

	go func() {
		for {
			select {
			case query := <-queryChan:
				g.menu.Search(query)
				menuLabel.SetText(matchCounterLabel())
				itemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
				resizeBasedOnResults()
			case items := <-g.menu.ItemsChan:
				g.menu.itemsMutex.Lock()
				deduplicated := make([]model.MenuItem, 0)
				seen := make(map[string]struct{})
				for _, item := range items {
					if _, ok := seen[item.Title]; !ok {
						seen[item.Title] = struct{}{}
						deduplicated = append(deduplicated, item)
					}
				}
				g.menu.items = deduplicated
				g.menu.itemsMutex.Unlock()
				g.menu.Search(g.menu.query)
				menuLabel.SetText(matchCounterLabel())
				itemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
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
			searchEntry.Disable()
			g.Quit(0)
		case fyne.KeyEscape:
			g.Quit(1)
		default:
			return
		}
		itemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
	}

	searchEntry.OnKeyDown = keyHandler
	myWindow.Canvas().SetOnTypedKey(keyHandler)

	myWindow.Resize(fyne.NewSize(g.dims.MinWidth, g.dims.MinHeight))
	mainContainer.Add(itemsCanvas.Container)
	myWindow.Canvas().Focus(searchEntry)
	myWindow.Show()
}

func createPidFile(name string) (string, error) {
	dir := os.TempDir()
	pidFile := fmt.Sprintf("%s/%s.pid", dir, name)
	if _, err := os.Stat(pidFile); err == nil {
		fmt.Println("Another instance of gmenu is already running. Exiting.")
		fmt.Println("If this is not the case, please delete the pid file:", pidFile)
		return "", fmt.Errorf("pid file already exists")

	}
	f, err := os.Create(pidFile)
	if err != nil {
		fmt.Println("Failed to create pid file")
		if ferr := f.Close(); ferr != nil {
			fmt.Println("Failed to close pid file:", ferr)
		}
		return "", err
	}
	return pidFile, f.Close()
}
