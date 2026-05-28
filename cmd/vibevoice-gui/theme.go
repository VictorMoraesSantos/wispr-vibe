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
		return color.NRGBA{R: 22, G: 22, B: 26, A: 255}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 225, G: 225, B: 230, A: 255}
	case theme.ColorNameButton:
		return color.NRGBA{R: 38, G: 38, B: 44, A: 255}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 32, G: 32, B: 36, A: 255}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 90, G: 90, B: 100, A: 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 99, G: 102, B: 241, A: 255} // indigo accent
	case theme.ColorNameFocus:
		return color.NRGBA{R: 99, G: 102, B: 241, A: 80}
	case theme.ColorNameHover:
		return color.NRGBA{R: 50, G: 50, B: 58, A: 255}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 30, G: 30, B: 35, A: 255}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 55, G: 55, B: 65, A: 255}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 110, G: 110, B: 125, A: 255}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 42, G: 42, B: 48, A: 255}
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 18, G: 18, B: 22, A: 240}
	case theme.ColorNameHeaderBackground:
		return color.NRGBA{R: 26, G: 26, B: 30, A: 255}
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
		return 10
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 18
	case theme.SizeNameSubHeadingText:
		return 15
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameInputRadius:
		return 6
	default:
		return theme.DarkTheme().Size(name)
	}
}
