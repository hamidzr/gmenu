package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type MenuItem struct {
	Title string
}

/*

TODO: add cli support for same behavior but in the terminal

*/

type SearchMethod func(MenuItem, string) bool

func directSearch(item MenuItem, keyword string) bool {
	return strings.Contains(strings.ToLower(item.Title), strings.ToLower(keyword))
}

type Menu struct {
	Items        []MenuItem
	Filtered     []MenuItem
	Selected     int
	ResultText   string
	SearchMethod SearchMethod
}

func NewMenu(itemTitles []string) Menu {
	items := make([]MenuItem, len(itemTitles))
	for i, entry := range itemTitles {
		items[i] = MenuItem{Title: entry}
	}
	if len(items) == 0 {
		panic("Menu must have at least one item")
	}
	return Menu{Items: items, Filtered: items,
		Selected:     0,
		SearchMethod: directSearch,
	}
}

func (m *Menu) Search(keyword string) {
	m.Filtered = nil // Reset the filtered list
	for _, item := range m.Items {
		if m.SearchMethod(item, keyword) {
			m.Filtered = append(m.Filtered, item)
		}
	}
	if len(m.Filtered) > 0 {
		m.Selected = 0
	}
	m.UpdateResultText()
}

// UpdateResultText creates a string representation of the filtered menu.
func (m *Menu) UpdateResultText() {
	m.ResultText = "\n"
	for i, item := range m.Filtered {
		if i == m.Selected {
			m.ResultText += fmt.Sprintf("-> %s\n", item.Title)
		} else {
			m.ResultText += fmt.Sprintf("   %s\n", item.Title)
		}
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
	var myWindow fyne.Window

	if deskDriver, ok := myApp.Driver().(desktop.Driver); ok {
		myWindow = deskDriver.CreateSplashWindow()
	} else {
		myWindow = myApp.NewWindow("choose")
	}

	searchEntry := &CustomEntry{}
	searchEntry.ExtendBaseWidget(searchEntry)
	searchEntry.SetPlaceHolder("Search")

	resultLabel := widget.NewLabel("")
	resultLabel.Wrapping = fyne.TextWrapWord

	menu.UpdateResultText()
	resultLabel.SetText(menu.ResultText)

	searchEntry.OnChanged = func(text string) {
		menu.Search(text)
		resultLabel.SetText(menu.ResultText)
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
			if menu.Selected >= 0 && menu.Selected < len(menu.Filtered) {
				fmt.Fprintln(os.Stdout, menu.Filtered[menu.Selected].Title)
				myApp.Quit()
			}
		case fyne.KeyEscape:
			os.Exit(1)
		}
		menu.UpdateResultText()
		resultLabel.SetText(menu.ResultText)
	}
	searchEntry.onKeyDown = keyHandler
	myWindow.Canvas().SetOnTypedKey(keyHandler)

	menuLabel := widget.NewLabel("Matched Items:")
	contentContainer := container.NewBorder(nil, nil, nil, nil, menuLabel, resultLabel)

	myWindow.SetContent(container.NewVBox(searchEntry, contentContainer))
	myWindow.Resize(fyne.NewSize(400, 300))     // Adjust as needed
	myWindow.SetOnClosed(func() { os.Exit(0) }) // Ensure the application exits properly

	// Set focus to the search entry on startup
	myWindow.Show()
	searchEntry.FocusGained()
	myWindow.Canvas().Focus(searchEntry)

	myWindow.ShowAndRun()
}
