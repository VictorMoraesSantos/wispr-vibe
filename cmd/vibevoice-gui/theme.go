package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type vibeTheme struct{}

func newVibeTheme() fyne.Theme {
	return &vibeTheme{}
}

func (v *vibeTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 20, G: 20, B: 25, A: 255}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 222, G: 220, B: 228, A: 255}
	case theme.ColorNameButton:
		return color.NRGBA{R: 34, G: 34, B: 40, A: 255}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 28, G: 28, B: 33, A: 255}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 82, G: 80, B: 92, A: 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 129, G: 110, B: 240, A: 255} // warm violet
	case theme.ColorNameFocus:
		return color.NRGBA{R: 129, G: 110, B: 240, A: 70}
	case theme.ColorNameHover:
		return color.NRGBA{R: 44, G: 42, B: 52, A: 255}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 26, G: 26, B: 32, A: 255}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 50, G: 48, B: 60, A: 255}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 100, G: 98, B: 112, A: 255}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 38, G: 36, B: 44, A: 255}
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 16, G: 16, B: 20, A: 245}
	case theme.ColorNameHeaderBackground:
		return color.NRGBA{R: 24, G: 24, B: 29, A: 255}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 60, G: 58, B: 70, A: 180}
	default:
		return theme.DarkTheme().Color(name, variant)
	}
}

func (v *vibeTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DarkTheme().Font(style)
}

func (v *vibeTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DarkTheme().Icon(name)
}

func (v *vibeTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 12
	case theme.SizeNameText:
		return 13
	case theme.SizeNameHeadingText:
		return 17
	case theme.SizeNameSubHeadingText:
		return 14
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameInputRadius:
		return 8
	default:
		return theme.DarkTheme().Size(name)
	}
}
