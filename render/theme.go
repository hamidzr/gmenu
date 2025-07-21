package render

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type MainTheme struct {
	fyne.Theme
}

func defaultThemeSizes() map[fyne.ThemeSizeName]float32 {
	sizes := map[fyne.ThemeSizeName]float32{
		theme.SizeNameInlineIcon:         float32(18),
		theme.SizeNameInnerPadding:       float32(6),
		theme.SizeNameLineSpacing:        float32(4),
		theme.SizeNamePadding:            float32(4),
		theme.SizeNameScrollBar:          float32(10),
		theme.SizeNameScrollBarSmall:     float32(2),
		theme.SizeNameSeparatorThickness: float32(1),
		theme.SizeNameText:               float32(16),
		theme.SizeNameHeadingText:        float32(30.6),
		theme.SizeNameSubHeadingText:     float32(24),
		theme.SizeNameCaptionText:        float32(14),
		theme.SizeNameInputBorder:        float32(2),
	}
	return sizes
}

func (m MainTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	default:
		return defaultThemeSizes()[name]
	}
}

func (m MainTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0x3d, G: 0x9f, B: 0xff, A: 0xff} // brighter blue
	case theme.ColorNameForeground:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0xe0, G: 0xe0, B: 0xe0, A: 0xff} // lighter text
		}
		return color.NRGBA{R: 0x20, G: 0x20, B: 0x20, A: 0xff}
	case theme.ColorNameBackground:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff} // darker background
		}
		return color.NRGBA{R: 0xf8, G: 0xf8, B: 0xf8, A: 0xff}
	case theme.ColorNameInputBackground:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0x2d, G: 0x2d, B: 0x2d, A: 0xff}
		}
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	case theme.ColorNamePlaceHolder:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff} // more visible placeholder
		}
		return color.NRGBA{R: 0x90, G: 0x90, B: 0x90, A: 0xff}
	case theme.ColorNameDisabled:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0x60, G: 0x60, B: 0x60, A: 0xff} // better contrast for counter
		}
		return color.NRGBA{R: 0xa0, G: 0xa0, B: 0xa0, A: 0xff}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}
