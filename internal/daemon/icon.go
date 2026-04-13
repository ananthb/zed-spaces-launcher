package daemon

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:embed assets/icon.png
var iconIdlePNG []byte

//go:embed assets/icon_active.png
var iconActivePNG []byte

//go:embed assets/icon_active.svg
var iconAppSVG []byte

// trayIconIdle returns the hollow cloud icon (no tracked codespaces).
// Uses ThemedResource so Fyne calls SetTemplateIcon on macOS, enabling
// automatic light/dark mode switching.
func trayIconIdle() fyne.Resource {
	return theme.NewThemedResource(fyne.NewStaticResource("icon.png", iconIdlePNG))
}

// trayIconActive returns the filled cloud icon (tracking codespaces).
func trayIconActive() fyne.Resource {
	return theme.NewThemedResource(fyne.NewStaticResource("icon_active.png", iconActivePNG))
}

// appIcon returns the app icon for dock/taskbar.
func appIcon() fyne.Resource {
	return fyne.NewStaticResource("icon.svg", iconAppSVG)
}
