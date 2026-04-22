// Package daemon — Cosmonaut custom Fyne theme.
//
// Palette and type scale derived from the Cosmonaut design system
// (Zed-inspired dark: graphite surfaces, lime accent, mono labels).
package daemon

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// cosmoTheme implements fyne.Theme with Cosmonaut's visual tokens.
// Colors are locked to the dark variant — the app presents consistently
// regardless of OS appearance. macOS tray icons still respect light/dark
// via SetTemplateIcon because they use theme.NewThemedResource.
type cosmoTheme struct{}

// Tokens — keep in sync with design-system (ui-primitives.jsx).
var (
	cBg           = color.NRGBA{0x0b, 0x0c, 0x0f, 0xff} // app background
	cBgAlt        = color.NRGBA{0x0f, 0x10, 0x12, 0xff} // sidebar / input chrome
	cSurface      = color.NRGBA{0x16, 0x17, 0x1a, 0xff} // cards, panels
	cSurface2     = color.NRGBA{0x1b, 0x1c, 0x1f, 0xff} // hover / selected row
	cBorder       = color.NRGBA{0x27, 0x27, 0x2a, 0xff} // 1px dividers
	cBorderStrong = color.NRGBA{0x3f, 0x3f, 0x46, 0xff}

	cText     = color.NRGBA{0xfa, 0xfa, 0xf9, 0xff}
	cTextDim  = color.NRGBA{0xa1, 0xa1, 0xaa, 0xff}
	cTextMute = color.NRGBA{0x71, 0x71, 0x7a, 0xff}
	cTextFain = color.NRGBA{0x52, 0x52, 0x5b, 0xff}

	cLime   = color.NRGBA{0xa3, 0xe6, 0x35, 0xff}
	cOrange = color.NRGBA{0xf9, 0x73, 0x16, 0xff}
	cRed    = color.NRGBA{0xef, 0x44, 0x44, 0xff}
	cBlue   = color.NRGBA{0x60, 0xa5, 0xfa, 0xff}
)

func (cosmoTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return cBg
	case theme.ColorNameForeground:
		return cText
	case theme.ColorNameForegroundOnPrimary:
		return cBg
	case theme.ColorNameDisabled, theme.ColorNameDisabledButton:
		return cTextFain
	case theme.ColorNameButton:
		return cSurface2
	case theme.ColorNameInputBackground, theme.ColorNameInputBorder:
		return cBgAlt
	case theme.ColorNamePlaceHolder:
		return cTextMute
	case theme.ColorNamePrimary:
		return cLime
	case theme.ColorNameHover:
		return color.NRGBA{0x2a, 0x2b, 0x2f, 0xff} // lighter than surface2 for visible hover
	case theme.ColorNamePressed:
		return color.NRGBA{0x33, 0x34, 0x38, 0xff}
	case theme.ColorNameFocus:
		return color.NRGBA{0xa3, 0xe6, 0x35, 0x40}
	case theme.ColorNameSelection:
		return color.NRGBA{0xa3, 0xe6, 0x35, 0x33}
	case theme.ColorNameSeparator:
		return cBorder
	case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return cBgAlt
	case theme.ColorNameScrollBar:
		return cBorderStrong
	case theme.ColorNameShadow:
		return color.NRGBA{0, 0, 0, 0x55}
	case theme.ColorNameError:
		return cRed
	case theme.ColorNameWarning:
		return cOrange
	case theme.ColorNameSuccess:
		return cLime
	case theme.ColorNameHyperlink:
		return cBlue
	}
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (cosmoTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Font returns Inter for UI text. Falls back to system default if the
// user hasn't installed Inter locally; monospace stays Fyne's default.
func (cosmoTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (cosmoTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 13
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameHeadingText:
		return 18
	case theme.SizeNameSubHeadingText:
		return 15
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 8
	case theme.SizeNameInlineIcon:
		return 14
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameScrollBar:
		return 10
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameInputRadius, theme.SizeNameSelectionRadius:
		return 5
	}
	return theme.DefaultTheme().Size(name)
}

// newCosmoTheme returns the Cosmonaut theme. Install in Run() via
//
//	d.app.Settings().SetTheme(newCosmoTheme())
func newCosmoTheme() fyne.Theme { return cosmoTheme{} }
