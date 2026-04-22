package daemon

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/ananth/cosmonaut/internal/codespace"
)

// repoPickerScreen shows a filterable list of repositories.
type repoPickerScreen struct {
	daemon   *Daemon
	win      fyne.Window
	onSelect func(repo string)
	onCancel func()

	allRepos    []string
	recentCount int
	filtered    []int
	filter      string

	list   *widget.List
	entry  *widget.Entry
	useAs  *widget.Button // "use as-is" button
	canvas fyne.CanvasObject
}

func (d *Daemon) newRepoPickerScreen(win fyne.Window, onSelect func(string), onCancel func()) *repoPickerScreen {
	s := &repoPickerScreen{
		daemon:   d,
		win:      win,
		onSelect: onSelect,
		onCancel: onCancel,
	}

	// Build initial repo list from cached codespaces + config.
	repos := codespace.UniqueRepos(d.Codespaces())
	repos = mergeRepos(repos, configRepos(d.Cfg))

	// Sort by history recency.
	hist := historyLoad()
	sorted := hist.SortRepos(repos)
	s.recentCount = countRecentRepos(sorted, hist)
	s.allRepos = sorted
	s.refilter()

	// Fetch all user repos in background to expand the list.
	go func() {
		allUserRepos, err := codespace.ListAllRepos(d.Runner)
		if err != nil {
			return
		}
		fyne.Do(func() {
			s.allRepos = mergeRepos(s.allRepos, allUserRepos)
			s.refilter()
			s.list.Refresh()
		})
	}()

	// Filter entry.
	s.entry = widget.NewEntry()
	s.entry.PlaceHolder = "Type to filter..."
	s.entry.OnChanged = func(text string) {
		s.filter = text
		s.refilter()
		s.list.Refresh()
		s.updateUseAs()
	}

	// Repo list.
	s.list = widget.NewList(
		func() int { return len(s.filtered) },
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			idx := s.filtered[id]
			label := s.allRepos[idx]
			if idx < s.recentCount {
				label += "  (recent)"
			}
			obj.(*widget.Label).SetText(label)
		},
	)
	s.list.OnSelected = func(id widget.ListItemID) {
		if id < len(s.filtered) {
			s.onSelect(s.allRepos[s.filtered[id]])
		}
		s.list.UnselectAll()
	}

	// "Use as-is" button for typed owner/repo.
	s.useAs = widget.NewButton("", func() {
		s.onSelect(s.filter)
	})
	s.useAs.Hidden = true

	cancelBtn := widget.NewButton("Cancel", func() {
		s.onCancel()
	})

	title := widget.NewLabel("Select a repository")
	title.TextStyle = fyne.TextStyle{Bold: true}

	s.canvas = container.NewBorder(
		container.NewVBox(title, s.entry),     // top
		container.NewVBox(s.useAs, cancelBtn), // bottom
		nil, nil,
		s.list, // center (fills remaining space)
	)

	return s
}

func (s *repoPickerScreen) refilter() {
	s.filtered = s.filtered[:0]
	lower := strings.ToLower(s.filter)
	for i, repo := range s.allRepos {
		if lower == "" || strings.Contains(strings.ToLower(repo), lower) {
			s.filtered = append(s.filtered, i)
		}
	}
}

func (s *repoPickerScreen) updateUseAs() {
	if strings.Contains(s.filter, "/") && !s.hasExactMatch() {
		s.useAs.SetText("Use \"" + s.filter + "\" as-is")
		s.useAs.Hidden = false
	} else {
		s.useAs.Hidden = true
	}
	s.useAs.Refresh()
}

func (s *repoPickerScreen) hasExactMatch() bool {
	lower := strings.ToLower(s.filter)
	for _, repo := range s.allRepos {
		if strings.ToLower(repo) == lower {
			return true
		}
	}
	return false
}
