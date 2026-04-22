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

//go:embed assets/icon_starting.png
var iconStartingPNG []byte

//go:embed assets/icon_error.png
var iconErrorPNG []byte

//go:embed assets/icon_active.svg
var iconAppSVG []byte

// trayIconIdle returns the helmet icon (no active codespaces).
func trayIconIdle() fyne.Resource {
	return theme.NewThemedResource(fyne.NewStaticResource("icon.png", iconIdlePNG))
}

// trayIconActive returns the helmet icon with filled dot (codespaces running).
func trayIconActive() fyne.Resource {
	return theme.NewThemedResource(fyne.NewStaticResource("icon_active.png", iconActivePNG))
}

// trayIconStarting returns the helmet icon with half dot (codespaces starting).
func trayIconStarting() fyne.Resource {
	return theme.NewThemedResource(fyne.NewStaticResource("icon_starting.png", iconStartingPNG))
}

// trayIconError returns the helmet icon with exclamation dot (error state).
func trayIconError() fyne.Resource {
	return theme.NewThemedResource(fyne.NewStaticResource("icon_error.png", iconErrorPNG))
}

// appIcon returns the app icon for dock/taskbar.
func appIcon() fyne.Resource {
	return fyne.NewStaticResource("icon.svg", iconAppSVG)
}
