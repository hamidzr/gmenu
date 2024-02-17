package render

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type MainTheme struct {
	fyne.Theme
}

func defaultThemeSizes() map[fyne.ThemeSizeName]float32 {
	sizes := map[fyne.ThemeSizeName]float32{
		theme.SizeNameInlineIcon:         float32(24),
		theme.SizeNameInnerPadding:       float32(10),
		theme.SizeNameLineSpacing:        float32(2),
		theme.SizeNamePadding:            float32(0),
		theme.SizeNameScrollBar:          float32(10),
		theme.SizeNameScrollBarSmall:     float32(2),
		theme.SizeNameSeparatorThickness: float32(1),
		theme.SizeNameText:               float32(16),
		theme.SizeNameHeadingText:        float32(30.6),
		theme.SizeNameSubHeadingText:     float32(24),
		theme.SizeNameCaptionText:        float32(15),
		theme.SizeNameInputBorder:        float32(3),
	}
	return sizes
}

func (m MainTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	default:
		return defaultThemeSizes()[name]
	}
}
