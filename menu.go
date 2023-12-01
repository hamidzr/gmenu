package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type MenuItem struct {
	ID    int
	Title string
}

type Menu struct {
	Items []MenuItem
}

func NewMenu() Menu {
	items := []MenuItem{
		{1, "Item 1"},
		{2, "Item 2"},
		{3, "Item 3"},
		{4, "Item 4"},
		{5, "Item 5"},
	}

	return Menu{Items: items}
}

func (m Menu) Search(keyword string) []MenuItem {
	var results []MenuItem
	for _, item := range m.Items {
		if strings.Contains(strings.ToLower(item.Title), strings.ToLower(keyword)) {
			results = append(results, item)
		}
	}
	return results
}

func main() {
	menu := NewMenu()

	myApp := app.New()
	myWindow := myApp.NewWindow("Menu")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search")

	resultLabel := widget.NewMultiLineEntry()
	resultLabel.Disable()

	// Update results as text changes
	searchEntry.OnChanged = func(text string) {
		results := menu.Search(text)

		resultText := ""
		for _, item := range results {
			resultText += fmt.Sprintf("[%d] %s\n", item.ID, item.Title)
		}

		resultLabel.SetText(resultText)
	}

	menuLabel := widget.NewLabel("Menu:")
	searchContainer := container.NewHBox(searchEntry)
	contentContainer := container.NewVBox(menuLabel, resultLabel)

	myWindow.SetContent(container.NewVBox(searchContainer, contentContainer))
	myWindow.ShowAndRun()
}
