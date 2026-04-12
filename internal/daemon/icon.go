package daemon

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed assets/icon.svg
var iconIdleSVG []byte

//go:embed assets/icon_active.svg
var iconActiveSVG []byte

// trayIconIdle returns the hollow cloud icon (no tracked codespaces).
func trayIconIdle() fyne.Resource {
	return fyne.NewStaticResource("icon.svg", iconIdleSVG)
}

// trayIconActive returns the filled cloud icon (tracking codespaces).
func trayIconActive() fyne.Resource {
	return fyne.NewStaticResource("icon_active.svg", iconActiveSVG)
}
