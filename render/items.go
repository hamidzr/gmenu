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

	// add icon if present
	var iconWidget *widget.Icon
	if item.Icon != "" {
		// simple icon mapping based on common patterns
		switch item.Icon {
		case "app", "application":
			iconWidget = widget.NewIcon(theme.ComputerIcon())
		case "file":
			iconWidget = widget.NewIcon(theme.DocumentIcon())
		case "folder", "directory":
			iconWidget = widget.NewIcon(theme.FolderIcon())
		default:
			iconWidget = widget.NewIcon(theme.InfoIcon())
		}
		iconWidget.Resize(fyne.NewSize(16, 16))
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

		// combine number, icon (if present), and text
		leftContent := container.NewHBox(numberContainer, spacer)
		if iconWidget != nil {
			leftContent.Add(iconWidget)
			iconSpacer := layout.NewSpacer()
			iconSpacer.Resize(fyne.NewSize(4, 1))
			leftContent.Add(iconSpacer)
		}

		textContent = container.NewBorder(nil, nil, leftContent, nil, optionText)
	} else {
		// just icon and text without numbers
		if iconWidget != nil {
			iconSpacer := layout.NewSpacer()
			iconSpacer.Resize(fyne.NewSize(4, 1))
			textContent = container.NewBorder(nil, nil, container.NewHBox(iconWidget, iconSpacer), nil, optionText)
		} else {
			textContent = container.NewStack(optionText)
		}
	}

	// create score metadata if needed
	var metadata *widget.Label
	if item.Score != 0 {
		metadata = widget.NewLabel(fmt.Sprintf("%d", item.Score))
		metadata.Alignment = fyne.TextAlignTrailing
		metadata.TextStyle = fyne.TextStyle{Bold: false, Italic: true}
	}

	// create background with better styling and separation
	background := canvas.NewRectangle(color.Transparent)
	if selected {
		background.FillColor = theme.Color(theme.ColorNameSelection)
	} else {
		// subtle alternating row colors for better item separation
		if idx%2 == 0 {
			background.FillColor = color.NRGBA{R: 0, G: 0, B: 0, A: 12} // slightly more visible stripe
		} else {
			background.FillColor = color.Transparent
		}
		// add subtle bottom border for item separation
		background.StrokeColor = color.NRGBA{R: 128, G: 128, B: 128, A: 30}
		background.StrokeWidth = 0.5
	}

	// add light padding for comfortable spacing
	paddedContent := container.NewWithoutLayout(textContent)
	paddedContent.Move(fyne.NewPos(4, 2))

	// compose final container with balanced padding
	var itemContainer *fyne.Container
	if metadata != nil {
		paddedMetadata := container.NewWithoutLayout(metadata)
		paddedMetadata.Move(fyne.NewPos(-4, 2))
		itemContainer = container.NewStack(background, container.NewBorder(nil, nil, nil, paddedMetadata, paddedContent))
	} else {
		itemContainer = container.NewStack(background, paddedContent)
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
