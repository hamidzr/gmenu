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

type Menu struct {
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

func NewMenu(
	itemTitles []string,
	initValue string,
	searchMethod SearchMethod,
	preserveOrder bool,
) *Menu {
	m := Menu{Selected: 0,
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
func (m *Menu) Search(keyword string) {
	m.queryMutex.Lock()
	m.query = keyword
	m.queryMutex.Unlock()
	if keyword == "" {
		m.Filtered = m.items
	} else {
		m.Filtered = m.SearchMethod(m.items, keyword, m.preserveOrder, m.resultLimit)
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

func (m *Menu) titlesToMenuItem(titles []string) []model.MenuItem {
	items := make([]model.MenuItem, len(titles))
	for i, entry := range titles {
		items[i] = model.MenuItem{Title: entry}
	}
	return items
}

type GMenu struct {
	AppTitle string
	prompt   string
	menuID   string
	menu     *Menu
	app      fyne.App
	store    store.Store
	ExitCode int
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
	store, err := store.NewFileStore(
		store.CacheDir(),
		store.ConfigDir(),
	)
	if err != nil {
		return nil, err
	}
	lastEntry := ""
	if menuID != "" {
		cache, err := store.LoadCache(menuID)
		if err != nil {
			return nil, err
		}
		if canBeHighlighted(cache.LastEntry) {
			lastEntry = cache.LastEntry
		}

	}
	menu := NewMenu(initialItems, lastEntry, searchMethod, preserveOrder)
	g := &GMenu{
		prompt:   prompt,
		AppTitle: title,
		menuID:   menuID,
		ExitCode: unsetInt,
		menu:     menu,
		store:    store,
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
	defer func() {
		if pidFile != "" {
			os.Remove(pidFile)
		}
	}()
	if err != nil {
		g.Quit(1)
		return err
	}
	g.app.Run()
	return nil
}

// SetItems sets the items to be displayed in the menu.
func (g *GMenu) SetItems(items []string) {
	g.menu.ItemsChan <- g.menu.titlesToMenuItem(items)
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
}

func (g *GMenu) PrependItems(items []string) {
	g.addItems(items, false)
}

func (g *GMenu) AppendItems(items []string) {
	g.addItems(items, true)
}

func (g *GMenu) cacheSelectedVal(value string) error {
	if g.menuID == "" {
		// skip caching if menuID is not set.
		return nil
	}
	cache, err := g.store.LoadCache(g.menuID)
	if err != nil {
		return err
	}
	cache.SetLastEntry(value)
	err = g.store.SaveCache(g.menuID, cache)
	if err != nil {
		return err
	}
	return nil
}

// SelectedItem returns the selected item.
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
		err := g.cacheSelectedVal(selected)
		if err != nil {
			return "", err
		}
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
	g.app.Settings().SetTheme(render.MainTheme{theme.DefaultTheme()})

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
	}
	searchEntry := &CustomEntry{disabledKeys: entryDisabledKeys}
	searchEntry.ExtendBaseWidget(searchEntry)
	searchEntry.SetPlaceHolder(g.prompt)
	searchEntry.SetText(g.menu.query)
	if g.menu.query != "" {
		searchEntry.SelectAll()
	}
	mainContainer := container.NewVBox(searchEntry)
	myWindow.SetContent(mainContainer)

	itemsCanvas := render.NewItemsCanvas()
	itemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
	// show match items out of total item count "Matched Items: [10/10]"
	matchCounterLabel := func() string {
		return fmt.Sprintf("Matched Items: [%d/%d]", g.menu.MatchCount, len(g.menu.items))
	}
	menuLabel := widget.NewLabel(matchCounterLabel())

	searchEntry.OnChanged = func(text string) {
		queryChan <- text
	}

	go func() {
		for {
			select {
			case query := <-queryChan:
				g.menu.Search(query)
				menuLabel.SetText(matchCounterLabel())
				itemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
			case items := <-g.menu.ItemsChan:
				g.menu.itemsMutex.Lock()
				g.menu.items = items
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
		case fyne.KeyDown:
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

	searchEntry.onKeyDown = keyHandler
	myWindow.Canvas().SetOnTypedKey(keyHandler)

	resultsContainer := container.NewBorder(nil, nil, nil, nil, menuLabel, itemsCanvas.Container)
	mainContainer.Add(resultsContainer)
	myWindow.Resize(fyne.NewSize(800, 300))

	myWindow.Canvas().Focus(searchEntry)
	myWindow.Show()
}

// CustomEntry is a widget.Entry that captures certain key events.
type CustomEntry struct {
	widget.Entry
	onKeyDown    func(key *fyne.KeyEvent)
	disabledKeys map[fyne.KeyName]bool
}

// SelectAll selects all text in the entry.
func (e *CustomEntry) SelectAll() {
	// TODO: this cannot select anything outside non-alphanumeric characters.
	e.Entry.DoubleTapped(nil)
}

// TypedKey implements the fyne.TypedKeyReceiver interface.
func (e *CustomEntry) TypedKey(key *fyne.KeyEvent) {
	if e.onKeyDown != nil {
		e.onKeyDown(key)
	}
	if e.disabledKeys != nil {
		if e.disabledKeys[key.Name] {
			return
		}
	}
	e.Entry.TypedKey(key)
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
	defer f.Close()
	if err != nil {
		fmt.Println("Failed to create pid file")
		return "", err
	}
	return pidFile, nil
}
