// Cosmonaut native UI primitives.
//
// Small building blocks reused across the unified window, codespace
// detail, create flow, and settings: keeps widget wiring consistent.
package daemon

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ── stateDot ────────────────────────────────────────────────────────────
// A small filled circle indicating codespace state. Matches the dots
// used in tray menu and sidebar.
func stateDot(state string) *canvas.Circle {
	var c color.Color
	switch state {
	case "Available":
		c = cLime
	case "Starting":
		c = cOrange
	case "Stopped":
		c = cTextFain
	case "Error":
		c = cRed
	default:
		c = cTextFain
	}
	dot := canvas.NewCircle(c)
	dot.StrokeWidth = 0
	dot.Resize(fyne.NewSize(8, 8))
	return dot
}

// ── caption ─────────────────────────────────────────────────────────────
// Small uppercase monospace section headers (e.g. "SSH CONNECTION").
func caption(text string) *canvas.Text {
	t := canvas.NewText(text, cTextMute)
	t.TextSize = 10
	t.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	return t
}

// ── surfaceCard ─────────────────────────────────────────────────────────
// Wraps content in a subtle surface with a 1px border: replaces Fyne's
// default grey "card" style with something closer to the design.
func surfaceCard(content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(cSurface)
	bg.StrokeColor = cBorder
	bg.StrokeWidth = 1
	bg.CornerRadius = 6
	return container.NewStack(bg, container.NewPadded(content))
}

// ── metaCell ────────────────────────────────────────────────────────────
// A uniform "label / value" stat tile used in the codespace detail grid.
func metaCell(label, value string, mono bool) fyne.CanvasObject {
	lbl := canvas.NewText(label, cTextMute)
	lbl.TextSize = 10
	lbl.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}

	val := canvas.NewText(value, cText)
	val.TextSize = 13
	val.TextStyle = fyne.TextStyle{Monospace: mono}

	return container.NewPadded(container.NewVBox(lbl, val))
}

// ── primaryButton / destructiveButton ───────────────────────────────────
// Fyne's built-in button importance flags map our accent colors in.
func primaryButton(label string, onTap func()) *widget.Button {
	b := widget.NewButton(label, onTap)
	b.Importance = widget.HighImportance // uses theme.ColorNamePrimary (lime)
	return b
}

func destructiveButton(label string, onTap func()) *widget.Button {
	b := widget.NewButton(label, onTap)
	b.Importance = widget.DangerImportance
	return b
}

// ── sidebarRow ──────────────────────────────────────────────────────────
// A flexible row with optional leading dot + trailing count badge, used
// in both the repo tree and any flat list view.
func sidebarRow(leading fyne.CanvasObject, label string, trailing fyne.CanvasObject) fyne.CanvasObject {
	lbl := widget.NewLabel(label)
	lbl.Truncation = fyne.TextTruncateEllipsis

	parts := []fyne.CanvasObject{}
	if leading != nil {
		parts = append(parts, leading)
	}
	parts = append(parts, lbl, layout.NewSpacer())
	if trailing != nil {
		parts = append(parts, trailing)
	}
	return container.NewHBox(parts...)
}

// ── countBadge ──────────────────────────────────────────────────────────
// Monospace pill showing e.g. "3" next to a repo row.
func countBadge(n int) fyne.CanvasObject {
	bg := canvas.NewRectangle(cSurface)
	bg.CornerRadius = 3
	txt := canvas.NewText(intToStr(n), cTextMute)
	txt.TextSize = 10
	txt.TextStyle = fyne.TextStyle{Monospace: true}
	return container.NewStack(bg, container.NewPadded(txt))
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{'0' + byte(n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
