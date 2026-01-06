package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type VoiceTypeTheme struct{}

func (m VoiceTypeTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameBackground {
		return color.Transparent
	}
	if name == theme.ColorNamePrimary {
		return color.RGBA{R: 225, G: 90, B: 164, A: 255} // Signature Pink
	}
	if name == theme.ColorNameForeground {
		return color.Black
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
		return 0 // Total removal of padding for "window-less" feel
	}
	if name == theme.SizeNameText {
		return 12
	}
	return theme.DefaultTheme().Size(name)
}
