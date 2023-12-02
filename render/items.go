package render

import (
	"fmt"
	"gmenu/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

/*
render a list of items
*/

type ItemsCanvas struct {
	*widget.Label
}

func NewItemsCanvas() *ItemsCanvas {
	label := widget.NewLabel("")
	label.Wrapping = fyne.TextWrapWord

	return &ItemsCanvas{
		Label: label,
	}
}

func (c *ItemsCanvas) ItemText(item model.MenuItem) string {
	return item.Title
}

func (c *ItemsCanvas) Update(items []model.MenuItem, selected int) {
	curText := "\n"
	for i, item := range items {
		if i == selected {
			curText += fmt.Sprintf("-> %s\n", c.ItemText(item))
		} else {
			curText += fmt.Sprintf("   %s\n", c.ItemText(item))
		}
	}
	c.SetText(curText)
}
