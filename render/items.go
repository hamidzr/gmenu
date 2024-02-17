package render

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
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
	c.Container.Objects = nil // Clear current items

	for i, item := range items {
		text := c.ItemText(item)

		// Create a label for the item
		label := widget.NewLabel(text)
		label.Wrapping = fyne.TextTruncate // Ensure text fits within the container

		background := canvas.NewRectangle(theme.BackgroundColor())
		if i == selected {
			// Highlight the selected item
			label.TextStyle = fyne.TextStyle{Bold: true}
			background.FillColor = theme.PrimaryColor()
		} else {
			background.StrokeColor = color.RGBA{R: 40, G: 40, B: 40, A: 255}
			background.StrokeWidth = 1
		}

		// Create a container for the label with the background
		// Use MaxLayout to ensure the label fills the width
		itemContainer := container.NewMax(background, label)
		itemContainer.Layout = layout.NewMaxLayout()

		c.Container.Add(itemContainer)
	}

	c.Container.Add(layout.NewSpacer()) // Add a final spacer for consistent look
	c.Container.Refresh()               // Refresh the container to apply changes
}
