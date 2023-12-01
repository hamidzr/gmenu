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
	m.UpdateResultText()
}

func (m *Menu) UpdateResultText() {
	// set result text to the first filtered item if any
	if len(m.Filtered) > 0 {
		m.Selected = 0
	} else {
		m.ResultText = ""
	}

	m.ResultText = "\n"
	for i, item := range m.Filtered {
		if i == m.Selected {
			m.ResultText += fmt.Sprintf("-> [%d] %s\n", item.ID, item.Title)
		} else {
			m.ResultText += fmt.Sprintf("   [%d] %s\n", item.ID, item.Title)
		}
	}
}

func main() {
	menu := NewMenu()

	myApp := app.New()
	myWindow := myApp.NewWindow("Menu")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search")

	resultLabel := widget.NewLabel("")
	resultLabel.Wrapping = fyne.TextWrapWord

	searchEntry.OnChanged = func(text string) {
		menu.Search(text)
		resultLabel.SetText(menu.ResultText)
	}
	searchEntry.OnSubmitted = func(text string) {
		fmt.Println("Submitted", text)
	}
	// TODO: listen to arrow keys on searchEntry

	myWindow.Canvas().SetOnTypedKey(func(keyEvent *fyne.KeyEvent) {
		switch keyEvent.Name {
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
	})

	menuLabel := widget.NewLabel("Menu:")
	contentContainer := container.NewBorder(nil, nil, nil, nil, menuLabel, resultLabel)

	myWindow.SetContent(container.NewVBox(searchEntry, contentContainer))
	myWindow.Resize(fyne.NewSize(400, 300)) // Adjust as needed
	myWindow.ShowAndRun()
}
