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

func RenderItem(item model.MenuItem, idx int, selected bool, noNumericSelection bool) *fyne.Container {
	// Safety check for item
	title := item.ComputedTitle()
	if title == "" {
		title = "Empty Item" // Fallback for empty items
	}

	// create the main text content
	optionText := widget.NewLabel(title)
	optionText.Truncation = fyne.TextTruncateEllipsis
	if selected {
		optionText.TextStyle = fyne.TextStyle{Bold: true}
	}

	var textContent *fyne.Container
	if !noNumericSelection && idx < 9 {
		// create number hint on the left with fixed width to prevent overlap
		numberHint := widget.NewLabel(fmt.Sprintf("%d", idx+1))
		numberHint.Alignment = fyne.TextAlignLeading
		numberHint.TextStyle = fyne.TextStyle{Bold: false, Italic: true}
		numberHint.Importance = widget.MediumImportance

		// create a container for the number with fixed width
		numberContainer := container.NewStack(numberHint)
		numberContainer.Resize(fyne.NewSize(20, numberContainer.MinSize().Height)) // fixed width for numbers

		// add some spacing between number and text
		spacer := layout.NewSpacer()
		spacer.Resize(fyne.NewSize(4, 1)) // small spacer

		textContent = container.NewBorder(nil, nil, container.NewHBox(numberContainer, spacer), nil, optionText)
	} else {
		textContent = container.NewStack(optionText)
	}

	// create score metadata if needed
	var metadata *widget.Label
	if item.Score != 0 {
		metadata = widget.NewLabel(fmt.Sprintf("%d", item.Score))
		metadata.Alignment = fyne.TextAlignTrailing
		metadata.TextStyle = fyne.TextStyle{Bold: false, Italic: true}
	}

	// create background
	background := canvas.NewRectangle(theme.BackgroundColor())
	if selected {
		background.FillColor = theme.PrimaryColor()
	} else {
		background.StrokeColor = color.RGBA{R: 40, G: 40, B: 40, A: 255}
		background.StrokeWidth = 1
	}

	// compose final container
	var itemContainer *fyne.Container
	if metadata != nil {
		itemContainer = container.NewStack(background, container.NewBorder(nil, nil, nil, metadata, textContent))
	} else {
		itemContainer = container.NewStack(background, textContent)
	}

	itemContainer.Layout = layout.NewStackLayout()
	return itemContainer
}

// Render updates the container with items, highlighting the selected one.
func (c *ItemsCanvas) Render(items []model.MenuItem, selected int, noNumericSelection bool) {
	// Safety checks to prevent nil pointer dereferences
	if c == nil || c.Container == nil {
		return
	}

	c.Container.Objects = nil // Clear current items

	// Handle empty items array
	if len(items) == 0 {
		c.Container.Add(layout.NewSpacer()) // Add a final spacer for consistent look
		c.Container.Refresh()               // Refresh the container to apply changes
		return
	}

	// Ensure selected index is within bounds
	if selected < 0 || selected >= len(items) {
		selected = 0
	}

	for i, item := range items {
		// Safety check for item title
		if item.ComputedTitle() == "" {
			continue // Skip empty items
		}
		c.Container.Add(RenderItem(item, i, i == selected, noNumericSelection))
	}

	c.Container.Add(layout.NewSpacer()) // Add a final spacer for consistent look
	c.Container.Refresh()               // Refresh the container to apply changes
}
