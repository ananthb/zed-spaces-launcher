package daemon

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// progressScreen displays an infinite progress bar with a status message.
type progressScreen struct {
	bar    *widget.ProgressBarInfinite
	status *widget.Label
	canvas fyne.CanvasObject
}

func newProgressScreen(message string) *progressScreen {
	bar := widget.NewProgressBarInfinite()
	status := widget.NewLabel(message)
	status.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		layout.NewSpacer(),
		bar,
		status,
		layout.NewSpacer(),
	)

	return &progressScreen{
		bar:    bar,
		status: status,
		canvas: container.NewPadded(content),
	}
}

func (p *progressScreen) setStatus(msg string) {
	p.status.SetText(msg)
}
