package render

import (
	"fyne.io/fyne/v2"
)

// ProportionalLayout is a custom layout that gives a fixed width to the second item
// and allocates the remaining space to the first item.
type ProportionalLayout struct {
	labelWidth float32
}

// NewProportionalLayout creates a new instance of ProportionalLayout.
func NewProportionalLayout(labelWidth float32) *ProportionalLayout {
	return &ProportionalLayout{labelWidth: labelWidth}
}

// Layout is called to position the contained objects within the specified size.
func (l *ProportionalLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 2 {
		return // This layout is designed specifically for two objects
	}

	// Allocate fixed width to label and remaining width to searchEntry
	searchEntrySize := fyne.NewSize(size.Width-l.labelWidth, size.Height)
	labelSize := fyne.NewSize(l.labelWidth, size.Height)

	// Position the searchEntry and label within the container
	objects[0].Resize(searchEntrySize)
	objects[0].Move(fyne.NewPos(0, 0)) // searchEntry starts at the left edge

	objects[1].Resize(labelSize)
	objects[1].Move(fyne.NewPos(searchEntrySize.Width, 0)) // label positioned after searchEntry
}

// MinSize calculates the minimum size of a container that uses this layout.
func (l *ProportionalLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var minWidth, minHeight float32
	for _, o := range objects {
		min := o.MinSize()
		minWidth += min.Width
		if min.Height > minHeight {
			minHeight = min.Height
		}
	}

	// Ensure the minimum width considers the fixed width for label
	minWidth += l.labelWidth

	return fyne.NewSize(minWidth, minHeight)
}
