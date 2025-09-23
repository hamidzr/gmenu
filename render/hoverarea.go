package render

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// hoverableArea provides a transparent tappable surface without altering visuals.
type hoverableArea struct {
	widget.BaseWidget
	onTap   func()
	onHover func(bool)
}

func newHoverableArea(onTap func(), onHover func(bool)) *hoverableArea {
	area := &hoverableArea{
		onTap:   onTap,
		onHover: onHover,
	}
	area.ExtendBaseWidget(area)
	return area
}

func (h *hoverableArea) Tapped(_ *fyne.PointEvent) {
	if h.onTap != nil {
		h.onTap()
	}
}

func (h *hoverableArea) MouseIn(_ *desktop.MouseEvent) {
	if h.onHover != nil {
		h.onHover(true)
	}
}

func (h *hoverableArea) MouseOut() {
	if h.onHover != nil {
		h.onHover(false)
	}
}

func (h *hoverableArea) MouseMoved(_ *desktop.MouseEvent) {
}

func (h *hoverableArea) CreateRenderer() fyne.WidgetRenderer {
	rect := canvas.NewRectangle(color.Transparent)
	return &hoverableAreaRenderer{rect: rect}
}

type hoverableAreaRenderer struct {
	rect *canvas.Rectangle
}

func (r *hoverableAreaRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (r *hoverableAreaRenderer) Layout(size fyne.Size) {
	r.rect.Resize(size)
}

func (r *hoverableAreaRenderer) Refresh() {
	canvas.Refresh(r.rect)
}

func (r *hoverableAreaRenderer) Destroy() {}

func (r *hoverableAreaRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.rect}
}
