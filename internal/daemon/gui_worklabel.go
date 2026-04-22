package daemon

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// workLabelScreen prompts the user for a work label before creating a codespace.
type workLabelScreen struct {
	entry    *widget.Entry
	hint     *widget.Label
	onCreate func(label string)
	onCancel func()
	canvas   fyne.CanvasObject
}

func newWorkLabelScreen(onCreate func(string), onCancel func()) *workLabelScreen {
	s := &workLabelScreen{
		onCreate: onCreate,
		onCancel: onCancel,
	}

	title := widget.NewLabel("What work are you planning to do?")
	title.TextStyle = fyne.TextStyle{Bold: true}

	s.entry = widget.NewEntry()
	s.entry.PlaceHolder = "e.g., fix indexer health checks"

	s.hint = widget.NewLabel("")
	s.hint.Hidden = true

	createBtn := widget.NewButton("Create", func() {
		text := strings.TrimSpace(s.entry.Text)
		if text == "" {
			s.hint.SetText("Enter a short label so the codespace is easier to recognize later.")
			s.hint.Hidden = false
			s.hint.Refresh()
			return
		}
		s.onCreate(text)
	})

	cancelBtn := widget.NewButton("Cancel", func() {
		s.onCancel()
	})

	s.canvas = container.NewVBox(
		layout.NewSpacer(),
		title,
		s.entry,
		s.hint,
		container.NewHBox(cancelBtn, layout.NewSpacer(), createBtn),
		layout.NewSpacer(),
	)

	return s
}
