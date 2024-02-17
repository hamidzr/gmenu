package render

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hamidzr/gmenu/model"
)

/*
render a list of items
*/

// ItemsCanvas is a container for showing a list of items.
type ItemsCanvas struct {
	Container   *fyne.Container
	LengthLimit int
}

// NewItemsCanvas initializes ItemsCanvas with a container.
func NewItemsCanvas() *ItemsCanvas {
	// Create an empty container for items
	cont := container.NewVBox()
	return &ItemsCanvas{
		Container:   cont,
		LengthLimit: 999,
	}
}

// ItemText shortens item text if necessary.
func (c *ItemsCanvas) ItemText(item model.MenuItem) string {
	if len(item.Title) > c.LengthLimit {
		return item.Title[:c.LengthLimit] + "..."
	}
	return item.Title
}

// Render updates the container with items, highlighting the selected one.
func (c *ItemsCanvas) Render(items []model.MenuItem, selected int) {
	c.Container.Objects = nil

	for i, item := range items {
		label := widget.NewLabel(c.ItemText(item))
		if i == selected {
			// Highlight the selected item
			label.TextStyle = fyne.TextStyle{Bold: true}
			background := canvas.NewRectangle(theme.PrimaryColor())
			background.FillColor = theme.PrimaryColor()
			border := container.NewWithoutLayout(background, label)
			background.Resize(border.Size())
			c.Container.Add(border)
		} else {
			c.Container.Add(label)
		}
	}

	c.Container.Refresh()
}
