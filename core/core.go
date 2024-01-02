package core

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/hamidzr/gmenu/model"
	"github.com/hamidzr/gmenu/render"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/sahilm/fuzzy"
)

type SearchMethod func([]model.MenuItem, string) []model.MenuItem

func isDirectMatch(item model.MenuItem, keyword string) bool {
	return strings.Contains(strings.ToLower(item.Title), strings.ToLower(keyword))
}
func directSearch(items []model.MenuItem, keyword string) []model.MenuItem {
	matches := make([]model.MenuItem, 0)
	for _, item := range items {
		if isDirectMatch(item, keyword) {
			matches = append(matches, item)
		}
	}
	return matches
}

func fuzzySearch(items []model.MenuItem, keyword string) []model.MenuItem {
	entries := make([]string, len(items))
	for i, item := range items {
		entries[i] = item.Title
	}

	matches := fuzzy.Find(keyword, entries)

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	result := make([]model.MenuItem, 0)
	for _, match := range matches {
		result = append(result, items[match.Index])
	}

	return result
}

type Menu struct {
	Items      []model.MenuItem
	Query      string
	itemsMutex sync.Mutex
	queryMutex sync.Mutex

	Filtered []model.MenuItem
	// zero-based index of the selected item in the filtered list
	Selected int
	// ResultText   string
	// MatchCount is the number of items that matched the search query.
	MatchCount   int
	SearchMethod SearchMethod
	resultLimit  int
}

func NewMenu(itemTitles []string) Menu {
	m := Menu{Selected: 0,
		SearchMethod: fuzzySearch,
		resultLimit:  10,
	}
	items := m.titlesToMenuItem(itemTitles)
	m.Items = items

	if len(items) == 0 {
		panic("Menu must have at least one item")
	}

	m.Search("")
	return m
}

// Filters the menu filtered list to only include items that match the keyword.
func (m *Menu) Search(keyword string) {
	if keyword == "" {
		m.Filtered = m.Items
	} else {
		m.Filtered = m.SearchMethod(m.Items, keyword)
	}
	if len(m.Filtered) > 0 {
		m.Selected = 0
	} else {
		m.Selected = -1
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
	menu     *Menu
	app      fyne.App
	exitCode int
}

// NewGMenu creates a new GMenu instance.
func NewGMenu(initialItems []string) *GMenu {
	menu := NewMenu(initialItems)
	g := &GMenu{
		exitCode: -1,
	}
	g.menu = &menu
	g.setupUI()
	return g
}

// Run starts the application.
func (g *GMenu) Run() {
	g.app.Run()
}

// SelectedItem returns the selected item.
func (g *GMenu) SelectedValue() (string, error) {
	// TODO: check if the app is running. using the doneChan?
	if g.exitCode == -1 {
		return "", fmt.Errorf("gmenu has not exited yet")
	}
	// TODO: cli option for allowing query.
	if g.exitCode != 0 {
		return "", fmt.Errorf("gmenu exited with code %d", g.exitCode)
	}
	if g.menu.Selected >= 0 && g.menu.Selected < len(g.menu.Filtered)+1 {
		return g.menu.Filtered[g.menu.Selected].Title, nil
	}
	return g.menu.Query, nil
}

// setupUI creates the UI elements.
func (g *GMenu) setupUI() {
	queryChan := make(chan string)
	itemsChan := make(chan []model.MenuItem)
	doneChan := make(chan bool)

	appTitle := "gmenu"
	myApp := app.New()
	g.app = myApp
	myApp.Settings().SetTheme(render.MainTheme{theme.DefaultTheme()})

	var myWindow fyne.Window
	if deskDriver, ok := myApp.Driver().(desktop.Driver); ok {
		myWindow = deskDriver.CreateSplashWindow()
	} else {
		myWindow = myApp.NewWindow("")
	}
	myWindow.SetTitle(appTitle)
	pidFile := createPidFile(appTitle)

	searchEntry := &CustomEntry{}
	searchEntry.ExtendBaseWidget(searchEntry)
	searchEntry.SetPlaceHolder("Search")
	mainContainer := container.NewVBox(searchEntry)
	myWindow.SetContent(mainContainer)

	itemsCanvas := render.NewItemsCanvas()
	itemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
	// show match items out of total item count "Matched Items: [10/10]"
	matchCounterLabel := func() string {
		return fmt.Sprintf("Matched Items: [%d/%d]", g.menu.MatchCount, len(g.menu.Items))
	}
	menuLabel := widget.NewLabel(matchCounterLabel())

	searchEntry.OnChanged = func(text string) {
		queryChan <- text
	}

	go func() {
		for {
			select {
			case query := <-queryChan:
				g.menu.queryMutex.Lock()
				g.menu.Query = query
				g.menu.queryMutex.Unlock()
				g.menu.Search(query)
				menuLabel.SetText(matchCounterLabel())
				itemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
			case items := <-itemsChan:
				g.menu.itemsMutex.Lock()
				g.menu.Items = items
				g.menu.itemsMutex.Unlock()
				g.menu.Search(g.menu.Query)
				menuLabel.SetText(matchCounterLabel())
				itemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
			case <-doneChan:
				return
			default:
				// CHECK: should we?
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	keyHandler := func(key *fyne.KeyEvent) {
		switch key.Name {
		case fyne.KeyDown:
			if g.menu.Selected < len(g.menu.Filtered)-1 {
				g.menu.Selected++
			}
		case fyne.KeyUp:
			if g.menu.Selected > 0 {
				g.menu.Selected--
			}
		case fyne.KeyReturn:
			g.exitCode = 0
			myApp.Quit()
		case fyne.KeyEscape:
			g.exitCode = 1
			myApp.Quit()

		}
		itemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
	}
	searchEntry.onKeyDown = keyHandler
	myWindow.Canvas().SetOnTypedKey(keyHandler)

	resultsContainer := container.NewBorder(nil, nil, nil, nil, menuLabel, itemsCanvas.Label)
	mainContainer.Add(resultsContainer)
	myWindow.Resize(fyne.NewSize(800, 300))
	myWindow.SetOnClosed(func() {
		doneChan <- true
		if g.exitCode == -1 {
			g.exitCode = 0
		}
		os.Remove(pidFile)
	}) // Ensure the application exits properly

	// Set focus to the search entry on startup
	// searchEntry.FocusGained()
	myWindow.Canvas().Focus(searchEntry)
	myWindow.Show()
}

// CustomEntry is a widget.Entry that captures certain key events.
type CustomEntry struct {
	widget.Entry
	onKeyDown func(key *fyne.KeyEvent)
}

func (e *CustomEntry) TypedKey(key *fyne.KeyEvent) {
	if e.onKeyDown != nil {
		e.onKeyDown(key)
	}
	// Call the parent method to ensure regular input handling.
	e.Entry.TypedKey(key)
}

func createPidFile(name string) string {
	dir := os.TempDir()
	pidFile := fmt.Sprintf("%s/%s.pid", dir, name)
	if _, err := os.Stat(pidFile); err == nil {
		fmt.Println("Another instance of gmenu is already running. Exiting.")
		fmt.Println("If this is not the case, please delete the pid file:", pidFile)
		os.Exit(1)
	}
	f, err := os.Create(pidFile)
	if err != nil {
		fmt.Println("Failed to create pid file")
		os.Exit(1)
	}
	defer f.Close()
	return pidFile
}
