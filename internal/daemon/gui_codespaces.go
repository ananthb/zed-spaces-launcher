package daemon

import (
	"fmt"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/ananth/cosmonaut/internal/codespace"
	"github.com/ananth/cosmonaut/internal/config"
)

// codespaceScreen shows codespaces for a repo with select, delete, create, and back.
type codespaceScreen struct {
	daemon      *Daemon
	win         fyne.Window
	repo        string
	target      config.Target
	targetName  string
	codespaces  []codespace.Codespace
	recommended int // index of recommended codespace, or -1

	onSelect func(cs *codespace.Codespace) // nil means "create new"
	onBack   func()
	onCancel func()

	list      *widget.List
	deleteBtn *widget.Button
	selected  int
	canvas    fyne.CanvasObject
}

func (d *Daemon) newCodespaceScreen(
	win fyne.Window,
	repo string,
	target config.Target,
	targetName string,
	onSelect func(*codespace.Codespace),
	onBack func(),
	onCancel func(),
) *codespaceScreen {
	s := &codespaceScreen{
		daemon:      d,
		win:         win,
		repo:        repo,
		target:      target,
		targetName:  targetName,
		onSelect:    onSelect,
		onBack:      onBack,
		onCancel:    onCancel,
		selected:    -1,
		recommended: -1,
	}

	// Get codespaces for this repo from cache.
	all := d.Codespaces()
	s.codespaces = codespace.FilterByRepo(all, repo)

	// Sort: active first, then alphabetically.
	sort.Slice(s.codespaces, func(i, j int) bool {
		oi, oj := stateOrder(s.codespaces[i].State), stateOrder(s.codespaces[j].State)
		if oi != oj {
			return oi < oj
		}
		return csLabel(s.codespaces[i]) < csLabel(s.codespaces[j])
	})

	// Find recommended.
	matches := codespace.FindMatching(s.codespaces, &target)
	if len(matches) == 1 {
		for i, cs := range s.codespaces {
			if cs.Name == matches[0].Name {
				s.recommended = i
				break
			}
		}
	}

	s.buildUI()
	return s
}

func (s *codespaceScreen) buildUI() {
	title := widget.NewLabel(fmt.Sprintf("Codespaces for %s", s.repo))
	title.TextStyle = fyne.TextStyle{Bold: true}

	// List: codespaces + "create new" row at the end.
	totalRows := len(s.codespaces) + 1
	s.list = widget.NewList(
		func() int { return totalRows },
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id < len(s.codespaces) {
				cs := s.codespaces[id]
				text := fmt.Sprintf("%s %s", stateIcon(cs.State), csLabel(cs))
				if id == s.recommended {
					text += "  [matches config]"
				}
				label.SetText(text)
			} else {
				label.SetText("+ Create a new codespace")
			}
		},
	)

	s.list.OnSelected = func(id widget.ListItemID) {
		s.selected = id
		s.deleteBtn.Enable()
		if id >= len(s.codespaces) {
			s.deleteBtn.Disable()
		}
	}

	// Double-click to select: use OnSelected for single click for now.
	// Actually, let's add a Select button instead.
	selectBtn := widget.NewButton("Open", func() {
		if s.selected < 0 {
			return
		}
		if s.selected < len(s.codespaces) {
			cs := s.codespaces[s.selected]
			s.onSelect(&cs)
		} else {
			s.onSelect(nil) // create new
		}
	})

	backBtn := widget.NewButton("Back", func() {
		s.onBack()
	})

	s.deleteBtn = widget.NewButton("Delete", func() {
		if s.selected < 0 || s.selected >= len(s.codespaces) {
			return
		}
		cs := s.codespaces[s.selected]
		dialog.ShowConfirm("Delete codespace",
			fmt.Sprintf("Delete %s? This cannot be undone.", cs.Name),
			func(ok bool) {
				if !ok {
					return
				}
				go func() {
					_ = codespace.DeleteCodespace(s.daemon.Runner, cs.Name)
					fyne.Do(func() {
						s.codespaces = removeCS(s.codespaces, cs.Name)
						s.selected = -1
						s.deleteBtn.Disable()
						s.list.Refresh()
					})
				}()
			},
			s.win,
		)
	})
	s.deleteBtn.Disable()

	buttons := container.NewHBox(backBtn, selectBtn, s.deleteBtn)

	s.canvas = container.NewBorder(
		title,   // top
		buttons, // bottom
		nil, nil,
		s.list, // center
	)
}

func removeCS(codespaces []codespace.Codespace, name string) []codespace.Codespace {
	var result []codespace.Codespace
	for _, cs := range codespaces {
		if cs.Name != name {
			result = append(result, cs)
		}
	}
	return result
}
