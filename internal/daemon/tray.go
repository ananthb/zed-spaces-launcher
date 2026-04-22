package daemon

import (
	"fmt"
	"os/exec"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"

	"github.com/ananth/codespace-zed/internal/codespace"
	"github.com/ananth/codespace-zed/internal/config"
	"github.com/ananth/codespace-zed/internal/history"
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
				go d.showPopover(name)
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
				go d.showPopover(args)
			})
			if sub := d.codespaceSubmenu(entry.Repository, args); sub != nil {
				item.ChildMenu = sub
			}
			items = append(items, item)
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

	// Preferences.
	items = append(items, fyne.NewMenuItemSeparator())
	items = append(items, d.preferencesMenuItem())

	// Quit.
	items = append(items, fyne.NewMenuItemSeparator())
	items = append(items, fyne.NewMenuItem("Quit", func() {
		d.Stop()
	}))

	return fyne.NewMenu("codespace-zed", items...)
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
			go d.showPopover("--codespace", cs.Name, launchArgs)
		}))
	}

	if len(repoCS) > maxSubmenuCodespaces {
		items = append(items, fyne.NewMenuItemSeparator())
		items = append(items, fyne.NewMenuItem("Show all...", func() {
			go d.showPopover(launchArgs)
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

// preferencesMenuItem builds the Preferences submenu item.
// Placeholder — filled in by the preferences implementation.
func (d *Daemon) preferencesMenuItem() *fyne.MenuItem {
	item := fyne.NewMenuItem("Preferences", nil)
	item.ChildMenu = d.buildPreferencesMenu()
	return item
}

// buildPreferencesMenu is a stub that will be replaced in the preferences step.
func (d *Daemon) buildPreferencesMenu() *fyne.Menu {
	var items []*fyne.MenuItem

	// GitHub auth.
	items = append(items, d.authMenuItems()...)

	items = append(items, fyne.NewMenuItemSeparator())

	// Daemon settings.
	if d.Cfg != nil && d.Cfg.Daemon != nil {
		items = append(items, d.daemonSettingsItems()...)
		items = append(items, fyne.NewMenuItemSeparator())
	}

	// Default target settings.
	if d.Cfg != nil && d.Cfg.DefaultTarget != "" {
		if _, ok := d.Cfg.Targets[d.Cfg.DefaultTarget]; ok {
			items = append(items, d.targetSettingsItems()...)
			items = append(items, fyne.NewMenuItemSeparator())
		}
	}

	// Edit config file.
	configPath := d.ConfigPath
	items = append(items, fyne.NewMenuItem("Edit config file...", func() {
		go openFile(configPath)
	}))

	return fyne.NewMenu("", items...)
}

// authMenuItems returns menu items for GitHub auth status and login/logout.
func (d *Daemon) authMenuItems() []*fyne.MenuItem {
	authed := d.checkAuthStatus()

	if authed {
		status := fyne.NewMenuItem("GitHub: authenticated", nil)
		status.Disabled = true
		logout := fyne.NewMenuItem("Remove GitHub auth", func() {
			go func() {
				_, _ = d.Runner.Run([]string{"auth", "logout", "--hostname", "github.com", "--yes"})
				d.rebuildTrayMenu()
			}()
		})
		return []*fyne.MenuItem{status, logout}
	}

	status := fyne.NewMenuItem("GitHub: not authenticated", nil)
	status.Disabled = true
	login := fyne.NewMenuItem("Log in to GitHub...", func() {
		go d.showGHAuthLogin()
	})
	return []*fyne.MenuItem{status, login}
}

// checkAuthStatus returns true if gh is authenticated.
func (d *Daemon) checkAuthStatus() bool {
	err := codespace.EnsureGHAuth(d.Runner)
	return err == nil
}

var hotkeyActions = []string{"picker", "previous", "default"}
var pollIntervals = []string{"1m", "5m", "15m", "30m"}

// daemonSettingsItems returns menu items for daemon configuration toggles.
func (d *Daemon) daemonSettingsItems() []*fyne.MenuItem {
	daemon := d.Cfg.Daemon

	// Hotkey action cycle.
	current := daemon.HotkeyAction
	if current == "" {
		current = "picker"
	}
	hotkeyItem := fyne.NewMenuItem(fmt.Sprintf("Hotkey action: %s", current), func() {
		go func() {
			d.Cfg.Daemon.HotkeyAction = cycleValue(hotkeyActions, d.Cfg.Daemon.HotkeyAction, "picker")
			d.saveAndRebuild()
		}()
	})

	// Poll interval cycle.
	poll := daemon.PollInterval
	if poll == "" {
		poll = "5m"
	}
	pollItem := fyne.NewMenuItem(fmt.Sprintf("Poll interval: %s", poll), func() {
		go func() {
			d.Cfg.Daemon.PollInterval = cycleValue(pollIntervals, d.Cfg.Daemon.PollInterval, "5m")
			d.saveAndRebuild()
		}()
	})

	return []*fyne.MenuItem{hotkeyItem, pollItem}
}

var autoStopValues = []string{"off", "15m", "30m", "1h"}
var preWarmValues = []string{"off", "08:00", "09:00", "10:00"}

// targetSettingsItems returns menu items for default target settings.
func (d *Daemon) targetSettingsItems() []*fyne.MenuItem {
	targetName := d.Cfg.DefaultTarget
	t := d.Cfg.Targets[targetName]

	// Auto-stop cycle.
	autoStop := t.AutoStop
	if autoStop == "" {
		autoStop = "off"
	}
	autoStopItem := fyne.NewMenuItem(fmt.Sprintf("Auto-stop: %s", autoStop), func() {
		go func() {
			t := d.Cfg.Targets[targetName]
			next := cycleValue(autoStopValues, t.AutoStop, "off")
			if next == "off" {
				next = ""
			}
			t.AutoStop = next
			d.Cfg.Targets[targetName] = t
			d.saveAndRebuild()
		}()
	})

	// Pre-warm cycle.
	preWarm := t.PreWarm
	if preWarm == "" {
		preWarm = "off"
	}
	preWarmItem := fyne.NewMenuItem(fmt.Sprintf("Pre-warm: %s", preWarm), func() {
		go func() {
			t := d.Cfg.Targets[targetName]
			next := cycleValue(preWarmValues, t.PreWarm, "off")
			if next == "off" {
				next = ""
			}
			t.PreWarm = next
			d.Cfg.Targets[targetName] = t
			d.saveAndRebuild()
		}()
	})

	heading := fyne.NewMenuItem(fmt.Sprintf("Target: %s", targetName), nil)
	heading.Disabled = true

	return []*fyne.MenuItem{heading, autoStopItem, preWarmItem}
}

// cycleValue advances to the next value in a list, wrapping around.
func cycleValue(values []string, current, defaultVal string) string {
	if current == "" {
		current = defaultVal
	}
	for i, v := range values {
		if v == current {
			return values[(i+1)%len(values)]
		}
	}
	return values[0]
}

// saveAndRebuild persists config and rebuilds the tray menu.
func (d *Daemon) saveAndRebuild() {
	if err := config.SaveConfig(d.ConfigPath, d.Cfg); err != nil {
		fmt.Printf("error saving config: %v\n", err)
	}
	d.rebuildTrayMenu()
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
