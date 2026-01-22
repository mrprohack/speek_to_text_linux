package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type VoiceTypeTheme struct{}

func (m VoiceTypeTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Make everything transparent for a floating wave effect
	if name == theme.ColorNameBackground ||
		name == theme.ColorNameOverlayBackground ||
		name == theme.ColorNameShadow ||
		name == theme.ColorNameSeparator ||
		name == theme.ColorNameMenuBackground ||
		name == theme.ColorNameInputBackground ||
		name == theme.ColorNameButton ||
		name == theme.ColorNameDisabledButton ||
		name == theme.ColorNameHover {
		return color.Transparent
	}
	if name == theme.ColorNamePrimary {
		return color.RGBA{R: 0, G: 240, B: 255, A: 255} // Electric Cyan
	}
	if name == theme.ColorNameForeground {
		return color.White
	}
	if name == theme.ColorNameSuccess {
		return color.RGBA{R: 16, G: 185, B: 129, A: 255} // Emerald
	}
	if name == theme.ColorNameError {
		return color.RGBA{R: 239, G: 68, B: 68, A: 255} // Crimson
	}
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (m VoiceTypeTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m VoiceTypeTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m VoiceTypeTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNamePadding {
		return 8
	}
	if name == theme.SizeNameText {
		return 14
	}
	if name == theme.SizeNameInputRadius {
		return 22
	}
	if name == theme.SizeNameInlineIcon {
		return 20
	}
	return theme.DefaultTheme().Size(name)
}
