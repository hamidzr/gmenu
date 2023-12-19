package main

import (
	"bufio"
	"fmt"
	"gmenu/model"
	"gmenu/render"
	"os"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/sahilm/fuzzy"
)

// max number of rendered results.
const resultLimit = 10

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
	Items    []model.MenuItem
	Filtered []model.MenuItem
	// zero-based index of the selected item in the filtered list
	Selected     int
	Query        string
	ResultText   string
	SearchMethod SearchMethod
}

func NewMenu(itemTitles []string) Menu {
	items := make([]model.MenuItem, len(itemTitles))
	for i, entry := range itemTitles {
		items[i] = model.MenuItem{Title: entry}
	}
	if len(items) == 0 {
		panic("Menu must have at least one item")
	}

	m := Menu{Items: items,
		Selected:     0,
		SearchMethod: fuzzySearch,
	}
	m.Search("")
	return m
}

// Filters the menu filtered list to only include items that match the keyword.
func (m *Menu) Search(keyword string) {
	m.Query = keyword
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
	if len(m.Filtered) > resultLimit {
		m.Filtered = m.Filtered[:resultLimit]
	}
}

// SetItems sets the menu items again.
func (m *Menu) SetItems(itemTitles []string) {
	items := make([]model.MenuItem, len(itemTitles))
	for i, entry := range itemTitles {
		items[i] = model.MenuItem{Title: entry}
	}
	m.Items = items
	// TODO: sync and search
	m.Search(m.Query)
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

func readItems() []string {
	var items []string
	// Check if there is any input from stdin (piped text)
	info, _ := os.Stdin.Stat()
	if (info.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			items = append(items, line)
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "error reading standard input:", err)
			os.Exit(1)
		}
	}

	// Proceed only if there are items
	if len(items) == 0 {
		fmt.Fprintln(os.Stderr, "No items provided through standard input")
		os.Exit(1)
	}
	return items
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

func main() {
	exitCode := 0
	appTitle := "gmenu"
	menu := NewMenu(readItems())
	myApp := app.New()
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
	itemsCanvas.Render(menu.Filtered, menu.Selected)

	searchEntry.OnChanged = func(text string) {
		menu.Search(text)
		itemsCanvas.Render(menu.Filtered, menu.Selected)
	}
	keyHandler := func(key *fyne.KeyEvent) {
		switch key.Name {
		case fyne.KeyDown:
			if menu.Selected < len(menu.Filtered)-1 {
				menu.Selected++
			}
		case fyne.KeyUp:
			if menu.Selected > 0 {
				menu.Selected--
			}
		case fyne.KeyReturn:
			if menu.Selected >= 0 && menu.Selected < len(menu.Filtered)+1 {
				fmt.Fprintln(os.Stdout, menu.Filtered[menu.Selected].Title)
			} else {
				// TODO: cli option.
				fmt.Fprintln(os.Stdout, menu.Query)
			}
			myApp.Quit()
		case fyne.KeyEscape:
			exitCode = 1
			myApp.Quit()

		}
		itemsCanvas.Render(menu.Filtered, menu.Selected)
	}
	searchEntry.onKeyDown = keyHandler
	myWindow.Canvas().SetOnTypedKey(keyHandler)

	menuLabel := widget.NewLabel("Matched Items:")
	resultsContainer := container.NewBorder(nil, nil, nil, nil, menuLabel, itemsCanvas.Label)
	mainContainer.Add(resultsContainer)
	myWindow.Resize(fyne.NewSize(800, 300))
	myWindow.SetOnClosed(func() {
		os.Remove(pidFile)
		os.Exit(exitCode)
	}) // Ensure the application exits properly

	// Set focus to the search entry on startup
	// searchEntry.FocusGained()
	myWindow.Canvas().Focus(searchEntry)
	myWindow.Show()
	myApp.Run()
}
