package render

import (
	"fmt"
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
	Container *fyne.Container
	// LengthLimit int
}

// NewItemsCanvas initializes ItemsCanvas with a container.
func NewItemsCanvas() *ItemsCanvas {
	// Create an empty container for items
	cont := container.NewVBox()
	return &ItemsCanvas{
		Container: cont,
		// LengthLimit: 999,
	}
}

func RenderItem(item model.MenuItem, idx int, selected bool) *fyne.Container {
	conf := model.GetGlobalConfig()

	var textLabel string
	if !conf.NoNumericSelection {
		textLabel = fmt.Sprintf("%d: %s", idx+1, item.ComputedTitle())
	} else {
		textLabel = item.ComputedTitle()
	}
	optionText := widget.NewLabel(textLabel)
	optionText.Truncation = fyne.TextTruncateEllipsis

	var metadata *widget.Label
	if item.Score != 0 {
		metadata = widget.NewLabel(fmt.Sprintf("%d", item.Score))
		metadata.Alignment = fyne.TextAlignTrailing
		metadata.TextStyle = fyne.TextStyle{Bold: false, Italic: true}
	}

	background := canvas.NewRectangle(theme.BackgroundColor())
	if selected { // Highlight the selected item
		if metadata != nil {
			optionText.TextStyle = fyne.TextStyle{Bold: true}
		}
		background.FillColor = theme.PrimaryColor()
	} else {
		background.StrokeColor = color.RGBA{R: 40, G: 40, B: 40, A: 255}
		background.StrokeWidth = 1
	}

	var itemContainer *fyne.Container
	if metadata != nil {
		itemContainer = container.NewStack(background, container.NewBorder(nil, nil, nil, metadata, optionText))
	} else {
		itemContainer = container.NewStack(background, optionText)
	}
	itemContainer.Layout = layout.NewStackLayout()
	return itemContainer
}

// Render updates the container with items, highlighting the selected one.
func (c *ItemsCanvas) Render(items []model.MenuItem, selected int) {
	c.Container.Objects = nil // Clear current items

	for i, item := range items {
		c.Container.Add(RenderItem(item, i, i == selected))
	}

	c.Container.Add(layout.NewSpacer()) // Add a final spacer for consistent look
	c.Container.Refresh()               // Refresh the container to apply changes
}
