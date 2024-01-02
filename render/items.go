package render

import (
	"fmt"

	"github.com/hamidzr/gmenu/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

/*
render a list of items
*/

type ItemsCanvas struct {
	*widget.Label
	LengthLimit int
}

func NewItemsCanvas() *ItemsCanvas {
	label := widget.NewLabel("")
	// label.Wrapping = fyne.TextWrapWord
	label.Wrapping = fyne.TextTruncate

	return &ItemsCanvas{
		Label:       label,
		LengthLimit: 999,
	}
}

func (c *ItemsCanvas) ItemText(item model.MenuItem) string {
	if len(item.Title) > c.LengthLimit {
		return item.Title[:c.LengthLimit] + "..."
	}
	return item.Title
}

func (c *ItemsCanvas) Render(items []model.MenuItem, selected int) {
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
