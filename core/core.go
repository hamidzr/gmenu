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
	MatchCount   int
	SearchMethod SearchMethod
	resultLimit  int
}

func NewMenu(itemTitles []string) Menu {
	m := Menu{Selected: 0,
		SearchMethod: fuzzySearch,
		resultLimit:  10,
		ItemsChan:    make(chan []model.MenuItem),
	}
	items := m.titlesToMenuItem(itemTitles)
	m.items = items

	if len(items) == 0 {
		panic("Menu must have at least one item")
	}

	m.Search("")
	return m
}

// Filters the menu filtered list to only include items that match the keyword.
func (m *Menu) Search(keyword string) {
	if keyword == "" {
		m.Filtered = m.items
	} else {
		m.Filtered = m.SearchMethod(m.items, keyword)
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
	AppTitle string
	menu     *Menu
	app      fyne.App
	ExitCode int
}

// NewGMenu creates a new GMenu instance.
func NewGMenu(initialItems []string) *GMenu {
	menu := NewMenu(initialItems)
	g := &GMenu{
		AppTitle: "gmenu",
		ExitCode: -1,
	}
	g.menu = &menu
	g.setupUI()
	return g
}

// Run starts the application.
func (g *GMenu) Run() error {
	pidFile, err := createPidFile(g.AppTitle)
	if err != nil {
		g.Quit(1)
		return err
	}
	g.app.Run()
	defer os.Remove(pidFile)
	return nil
}

// SetItems sets the items to be displayed in the menu.
func (g *GMenu) SetItems(items []string) {
	g.menu.ItemsChan <- g.menu.titlesToMenuItem(items)
}

// SelectedItem returns the selected item.
func (g *GMenu) SelectedValue() (string, error) {
	// TODO: check if the app is running. using the doneChan?
	if g.ExitCode == -1 {
		return "", fmt.Errorf("gmenu has not exited yet")
	}
	// TODO: cli option for allowing query.
	if g.ExitCode != 0 {
		return "", fmt.Errorf("gmenu exited with code %d", g.ExitCode)
	}
	if g.menu.Selected >= 0 && g.menu.Selected < len(g.menu.Filtered)+1 {
		return g.menu.Filtered[g.menu.Selected].Title, nil
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

	var myWindow fyne.Window
	if deskDriver, ok := g.app.Driver().(desktop.Driver); ok {
		myWindow = deskDriver.CreateSplashWindow()
	} else {
		myWindow = g.app.NewWindow("")
	}
	myWindow.SetTitle(g.AppTitle)
	searchEntry := &CustomEntry{}
	searchEntry.ExtendBaseWidget(searchEntry)
	searchEntry.SetPlaceHolder("Search")
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
				g.menu.queryMutex.Lock()
				g.menu.query = query
				g.menu.queryMutex.Unlock()
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
			}
		case fyne.KeyUp:
			if g.menu.Selected > 0 {
				g.menu.Selected--
			}
		case fyne.KeyReturn:
			g.Quit(0)
		case fyne.KeyEscape:
			g.Quit(1)

		}
		itemsCanvas.Render(g.menu.Filtered, g.menu.Selected)
	}
	searchEntry.onKeyDown = keyHandler
	myWindow.Canvas().SetOnTypedKey(keyHandler)

	resultsContainer := container.NewBorder(nil, nil, nil, nil, menuLabel, itemsCanvas.Label)
	mainContainer.Add(resultsContainer)
	myWindow.Resize(fyne.NewSize(800, 300))

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
