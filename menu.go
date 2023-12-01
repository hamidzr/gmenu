package main

import (
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type MenuItem struct {
	ID    int
	Title string
}

type Menu struct {
	Items      []MenuItem
	Filtered   []MenuItem
	Selected   int
	ResultText string
}

func NewMenu() Menu {
	items := []MenuItem{
		{1, "Item pp 1"},
		{2, "Item pp 2"},
		{3, "Item abp 3"},
		{4, "Item gga 4"},
		{5, "Item 5"},
	}
	if len(items) == 0 {
		panic("Menu must have at least one item")
	}
	return Menu{Items: items, Filtered: items, Selected: 0}
}

func (m *Menu) Search(keyword string) {
	m.Filtered = nil // Reset the filtered list
	for _, item := range m.Items {
		if strings.Contains(strings.ToLower(item.Title), strings.ToLower(keyword)) {
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
			m.ResultText += fmt.Sprintf("-> [%d] %s\n", item.ID, item.Title)
		} else {
			m.ResultText += fmt.Sprintf("   [%d] %s\n", item.ID, item.Title)
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

func main() {
	menu := NewMenu()

	myApp := app.New()
	myWindow := myApp.NewWindow("Menu")

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
	// Implement arrow key navigation
	searchEntry.onKeyDown = func(key *fyne.KeyEvent) {
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
		}
		menu.UpdateResultText()
		resultLabel.SetText(menu.ResultText)
	}

	menuLabel := widget.NewLabel("Menu:")
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
