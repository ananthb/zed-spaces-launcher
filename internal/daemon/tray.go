package daemon

import (
	"fmt"
	"os/exec"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"

	"github.com/ananth/cosmonaut/internal/codespace"
	"github.com/ananth/cosmonaut/internal/history"
)

const maxSubmenuCodespaces = 5

// buildTrayMenu constructs the system tray menu from config, history,
// and cached codespace state.
func (d *Daemon) buildTrayMenu() *fyne.Menu {
	var items []*fyne.MenuItem
	seen := make(map[string]bool)

	// ── Spaces heading ──
	heading := fyne.NewMenuItem("Spaces", nil)
	heading.Disabled = true
	items = append(items, heading)

	// Default target.
	if d.Cfg != nil && d.Cfg.DefaultTarget != "" {
		if t, ok := d.Cfg.Targets[d.Cfg.DefaultTarget]; ok {
			name := d.Cfg.DefaultTarget
			item := fyne.NewMenuItem("Open "+name, func() {
				go d.showGUI(name)
			})
			if sub := d.codespaceSubmenu(t.Repository, name); sub != nil {
				item.ChildMenu = sub
			}
			items = append(items, item)
			seen[t.Repository] = true
		}
	}

	// Recent targets from history (de-duplicated against default).
	hist := history.Load()
	if len(hist.Entries) > 0 {
		items = append(items, fyne.NewMenuItemSeparator())
		limit := min(5, len(hist.Entries))
		for i := len(hist.Entries) - 1; i >= len(hist.Entries)-limit; i-- {
			entry := hist.Entries[i]
			if seen[entry.Repository] {
				continue
			}
			seen[entry.Repository] = true

			targetName := d.targetNameForRepo(entry.Repository)
			label := entry.Repository
			args := targetName
			if args == "" {
				args = entry.Repository
			}
			item := fyne.NewMenuItem(label, func() {
				go d.showGUI(args)
			})
			if sub := d.codespaceSubmenu(entry.Repository, args); sub != nil {
				item.ChildMenu = sub
			}
			items = append(items, item)
		}
	}

	// Open previous / launch.
	items = append(items, fyne.NewMenuItemSeparator())
	if len(hist.Entries) > 0 {
		items = append(items, fyne.NewMenuItem("Open previous", func() {
			go d.hotkeyActionPrevious()
		}))
	}
	items = append(items, fyne.NewMenuItem("Launch...", func() {
		go d.showGUI()
	}))

	// Preferences.
	items = append(items, fyne.NewMenuItemSeparator())
	items = append(items, d.preferencesMenuItem())

	// Quit.
	items = append(items, fyne.NewMenuItemSeparator())
	items = append(items, fyne.NewMenuItem("Quit", func() {
		d.Stop()
	}))

	return fyne.NewMenu("cosmonaut", items...)
}

// codespaceSubmenu builds a submenu showing codespaces for a repo.
// Returns nil if the repo has no codespaces.
func (d *Daemon) codespaceSubmenu(repo, launchArgs string) *fyne.Menu {
	all := d.Codespaces()
	repoCS := codespace.FilterByRepo(all, repo)
	if len(repoCS) == 0 {
		return nil
	}

	// Sort: Available/Starting first, then others, alphabetically within groups.
	sort.Slice(repoCS, func(i, j int) bool {
		oi, oj := stateOrder(repoCS[i].State), stateOrder(repoCS[j].State)
		if oi != oj {
			return oi < oj
		}
		return csLabel(repoCS[i]) < csLabel(repoCS[j])
	})

	var items []*fyne.MenuItem
	limit := min(maxSubmenuCodespaces, len(repoCS))
	for _, cs := range repoCS[:limit] {
		label := fmt.Sprintf("%s %s", stateIcon(cs.State), csLabel(cs))
		items = append(items, fyne.NewMenuItem(label, func() {
			go d.showGUI("--codespace", cs.Name, launchArgs)
		}))
	}

	if len(repoCS) > maxSubmenuCodespaces {
		items = append(items, fyne.NewMenuItemSeparator())
		items = append(items, fyne.NewMenuItem("Show all...", func() {
			go d.showGUI(launchArgs)
		}))
	}

	return fyne.NewMenu("", items...)
}

// stateOrder returns a sort key for codespace states (lower = first).
func stateOrder(state string) int {
	switch state {
	case "Available":
		return 0
	case "Starting":
		return 1
	case "Stopped":
		return 2
	default:
		return 3
	}
}

// stateIcon returns a Unicode indicator for a codespace state.
func stateIcon(state string) string {
	switch state {
	case "Available":
		return "●"
	case "Starting":
		return "◐"
	default:
		return "○"
	}
}

// csLabel returns a short display label for a codespace.
func csLabel(cs codespace.Codespace) string {
	name := cs.DisplayName
	if name == "" {
		name = cs.Name
	}
	if cs.GitStatus != nil {
		ref := cs.GitStatus.Ref
		if ref == "" {
			ref = cs.GitStatus.Branch
		}
		if ref != "" {
			return fmt.Sprintf("%s (%s)", name, ref)
		}
	}
	return name
}

// targetNameForRepo returns the config target name for a repo, or empty string.
func (d *Daemon) targetNameForRepo(repo string) string {
	if d.Cfg == nil {
		return ""
	}
	for name, t := range d.Cfg.Targets {
		if t.Repository == repo {
			return name
		}
	}
	return ""
}

// preferencesMenuItem opens the preferences window.
func (d *Daemon) preferencesMenuItem() *fyne.MenuItem {
	return fyne.NewMenuItem("Preferences...", func() {
		go d.showPreferences()
	})
}

// rebuildTrayMenu rebuilds and replaces the system tray menu.
// Safe to call from any goroutine.
func (d *Daemon) rebuildTrayMenu() {
	if d.app == nil {
		return
	}
	fyne.Do(func() {
		if desk, ok := d.app.(desktop.App); ok {
			desk.SetSystemTrayMenu(d.buildTrayMenu())
		}
	})
}

// openFile opens a file with the OS default handler.
func openFile(path string) {
	_ = exec.Command("open", path).Run()
}
