package daemon

import (
	"fyne.io/fyne/v2"

	"github.com/ananth/codespace-zed/internal/history"
)

// buildTrayMenu constructs the system tray menu from config and history.
func (d *Daemon) buildTrayMenu() *fyne.Menu {
	var items []*fyne.MenuItem

	// Default target.
	if d.Cfg != nil && d.Cfg.DefaultTarget != "" {
		if _, ok := d.Cfg.Targets[d.Cfg.DefaultTarget]; ok {
			name := d.Cfg.DefaultTarget
			items = append(items, fyne.NewMenuItem("Open "+name, func() {
				go d.showPopover(name)
			}))
		}
	}

	// Recent targets from history.
	hist := history.Load()
	if len(hist.Entries) > 0 {
		items = append(items, fyne.NewMenuItemSeparator())
		limit := min(5, len(hist.Entries))
		for i := len(hist.Entries) - 1; i >= len(hist.Entries)-limit; i-- {
			entry := hist.Entries[i]
			targetName := d.targetNameForRepo(entry.Repository)
			label := entry.Repository
			// Prefer config target name; fall back to repo name so the
			// child process can skip the repo selector TUI.
			args := targetName
			if args == "" {
				args = entry.Repository
			}
			items = append(items, fyne.NewMenuItem(label, func() {
				go d.showPopover(args)
			}))
		}
	}

	// Open previous / picker.
	items = append(items, fyne.NewMenuItemSeparator())
	if len(hist.Entries) > 0 {
		items = append(items, fyne.NewMenuItem("Open previous", func() {
			go d.hotkeyActionPrevious()
		}))
	}
	items = append(items, fyne.NewMenuItem("Open picker...", func() {
		go d.showPopover()
	}))

	// Quit.
	items = append(items, fyne.NewMenuItemSeparator())
	items = append(items, fyne.NewMenuItem("Quit", func() {
		d.Stop()
	}))

	return fyne.NewMenu("codespace-zed", items...)
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
