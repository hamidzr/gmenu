package main

import (
	"bufio"
	"fmt"
	"gmenu/model"
	"gmenu/render"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/lithammer/fuzzysearch/fuzzy"
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
	ranks := fuzzy.RankFindFold(keyword, entries)
	matches := make([]model.MenuItem, 0)
	for _, rank := range ranks {
		matches = append(matches, items[rank.OriginalIndex])
	}
	return matches
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
	m.Filtered = m.SearchMethod(m.Items, keyword)
	if len(m.Filtered) > 0 {
		m.Selected = 0
	} else {
		m.Selected = -1
	}
	if len(m.Filtered) > resultLimit {
		m.Filtered = m.Filtered[:resultLimit]
	}
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

func main() {
	menu := NewMenu(readItems())
	myApp := app.New()
	myApp.Settings().SetTheme(render.MainTheme{theme.DefaultTheme()})

	var myWindow fyne.Window
	if deskDriver, ok := myApp.Driver().(desktop.Driver); ok {
		myWindow = deskDriver.CreateSplashWindow()
	} else {
		myWindow = myApp.NewWindow("")
	}
	myWindow.SetTitle("gmenu")

	searchEntry := &CustomEntry{}
	searchEntry.ExtendBaseWidget(searchEntry)
	searchEntry.SetPlaceHolder("Search")

	itemsCanvas := render.NewItemsCanvas()
	itemsCanvas.Update(menu.Filtered, menu.Selected)

	searchEntry.OnChanged = func(text string) {
		menu.Search(text)
		itemsCanvas.Update(menu.Filtered, menu.Selected)
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
			os.Exit(1)
		}
		itemsCanvas.Update(menu.Filtered, menu.Selected)
	}
	searchEntry.onKeyDown = keyHandler
	myWindow.Canvas().SetOnTypedKey(keyHandler)

	menuLabel := widget.NewLabel("Matched Items:")
	contentContainer := container.NewBorder(nil, nil, nil, nil, menuLabel, itemsCanvas.Label)

	myWindow.SetContent(container.NewVBox(searchEntry, contentContainer))
	myWindow.Resize(fyne.NewSize(800, 300))
	myWindow.SetOnClosed(func() { os.Exit(0) }) // Ensure the application exits properly

	// Set focus to the search entry on startup
	// searchEntry.FocusGained()
	myWindow.Canvas().Focus(searchEntry)
	myWindow.ShowAndRun()
}
