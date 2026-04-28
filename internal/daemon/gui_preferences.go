package daemon

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/linuskendall/cosmonaut/internal/codespace"
	"github.com/linuskendall/cosmonaut/internal/config"
)

// buildSettingsPanel builds the settings content panel for the unified window.
func (d *Daemon) buildSettingsPanel(win fyne.Window) fyne.CanvasObject {
	var items []fyne.CanvasObject

	heading := widget.NewLabel("Settings")
	heading.TextStyle = fyne.TextStyle{Bold: true}
	items = append(items, heading)

	// GitHub auth section.
	items = append(items, d.buildAuthSection(win))
	items = append(items, widget.NewSeparator())

	// Editor selection.
	items = append(items, d.buildEditorSection())
	items = append(items, widget.NewSeparator())

	// Daemon settings.
	if d.Cfg != nil {
		if d.Cfg.Daemon == nil {
			d.Cfg.Daemon = &config.DaemonConfig{}
		}
		items = append(items, d.buildDaemonSection())
		items = append(items, widget.NewSeparator())
	}

	// Default target settings.
	if d.Cfg != nil && d.Cfg.DefaultTarget != "" {
		if _, ok := d.Cfg.Targets[d.Cfg.DefaultTarget]; ok {
			items = append(items, d.buildTargetSection())
			items = append(items, widget.NewSeparator())
		}
	}

	// Edit config file button.
	configPath := d.ConfigPath
	items = append(items, widget.NewButton("Edit config file...", func() {
		go openFile(configPath)
	}))

	return container.NewVScroll(container.NewPadded(container.NewVBox(items...)))
}

// showPreferences opens settings as a separate window (called from tray menu).
func (d *Daemon) showPreferences() {
	if d.app == nil {
		return
	}
	fyne.Do(func() {
		win := d.app.NewWindow("Cosmonaut Settings")
		win.Resize(fyne.NewSize(420, 400))
		win.SetFixedSize(true)
		win.CenterOnScreen()
		win.SetContent(d.buildSettingsPanel(win))
		win.Show()
	})
}

func (d *Daemon) buildAuthSection(win fyne.Window) fyne.CanvasObject {
	authed := codespace.EnsureGHAuth(d.Runner) == nil

	var statusText string
	if authed {
		statusText = "GitHub: authenticated"
	} else {
		statusText = "GitHub: not authenticated"
	}
	statusLabel := widget.NewLabel(statusText)

	// After an auth-state change, the section's button label and the tray
	// menu both need to reflect the new state. Rebuilding the whole settings
	// panel is the simplest correct refresh: the panel is small, all sections
	// re-read their state on construction, and there's no other live state
	// in the window worth preserving across an auth flip.
	refresh := func() {
		d.rebuildTrayMenu()
		if win != nil {
			win.SetContent(d.buildSettingsPanel(win))
		}
	}

	var actionBtn *widget.Button
	if authed {
		actionBtn = widget.NewButton("Remove auth", func() {
			actionBtn.Disable()
			go func() {
				_, err := d.Runner.Run([]string{"auth", "logout", "--hostname", "github.com", "--yes"})
				fyne.Do(func() {
					if err != nil {
						log.Printf("auth logout: %v", err)
						dialog.ShowError(fmt.Errorf("gh auth logout failed: %w", err), win)
						actionBtn.Enable()
						return
					}
					refresh()
				})
			}()
		})
	} else {
		actionBtn = widget.NewButton("Log in...", func() {
			actionBtn.Disable()
			go func() {
				_, err := d.Runner.Run([]string{"auth", "login", "--web", "--hostname", "github.com"})
				fyne.Do(func() {
					if err != nil {
						log.Printf("auth login: %v", err)
						dialog.ShowError(fmt.Errorf("gh auth login failed: %w", err), win)
						actionBtn.Enable()
						return
					}
					refresh()
				})
			}()
		})
	}

	return container.NewHBox(statusLabel, layout.NewSpacer(), actionBtn)
}

func (d *Daemon) buildEditorSection() fyne.CanvasObject {
	currentEditor := d.Cfg.Editor
	if currentEditor == "" {
		currentEditor = "zed"
	}
	editorSelect := widget.NewSelect([]string{"zed", "neovim"}, func(val string) {
		d.Cfg.Editor = val
		d.persistConfig()
	})
	editorSelect.Selected = currentEditor

	return widget.NewForm(widget.NewFormItem("Editor", editorSelect))
}

func (d *Daemon) buildDaemonSection() fyne.CanvasObject {
	daemon := d.Cfg.Daemon

	currentAction := daemon.HotkeyAction
	if currentAction == "" {
		currentAction = "picker"
	}
	actionSelect := widget.NewSelect([]string{"picker", "previous", "default"}, func(val string) {
		d.Cfg.Daemon.HotkeyAction = val
		d.persistConfig()
	})
	actionSelect.Selected = currentAction

	currentPoll := daemon.PollInterval
	if currentPoll == "" {
		currentPoll = "5m"
	}
	pollSelect := widget.NewSelect([]string{"1m", "5m", "15m", "30m"}, func(val string) {
		d.Cfg.Daemon.PollInterval = val
		d.persistConfig()
	})
	pollSelect.Selected = currentPoll

	currentInhibit := daemon.InhibitSleep
	if currentInhibit == "" {
		currentInhibit = "off"
	}
	inhibitSelect := widget.NewSelect([]string{"off", "sleep", "sleep+shutdown"}, func(val string) {
		d.Cfg.Daemon.InhibitSleep = val
		if d.sessions != nil {
			d.sessions.SetMode(val)
		}
		d.persistConfig()
	})
	inhibitSelect.Selected = currentInhibit

	return widget.NewForm(
		widget.NewFormItem("Hotkey action", actionSelect),
		widget.NewFormItem("Poll interval", pollSelect),
		widget.NewFormItem("Inhibit sleep", inhibitSelect),
	)
}

func (d *Daemon) buildTargetSection() fyne.CanvasObject {
	targetName := d.Cfg.DefaultTarget
	t := d.Cfg.Targets[targetName]

	heading := widget.NewLabel(fmt.Sprintf("Target: %s", targetName))
	heading.TextStyle = fyne.TextStyle{Bold: true}

	currentAutoStop := t.AutoStop
	if currentAutoStop == "" {
		currentAutoStop = "off"
	}
	autoStopSelect := widget.NewSelect([]string{"off", "15m", "30m", "1h"}, func(val string) {
		t := d.Cfg.Targets[targetName]
		if val == "off" {
			t.AutoStop = ""
		} else {
			t.AutoStop = val
		}
		d.Cfg.Targets[targetName] = t
		d.persistConfig()
	})
	autoStopSelect.Selected = currentAutoStop

	currentPreWarm := t.PreWarm
	if currentPreWarm == "" {
		currentPreWarm = "off"
	}
	preWarmSelect := widget.NewSelect([]string{"off", "08:00", "09:00", "10:00"}, func(val string) {
		t := d.Cfg.Targets[targetName]
		if val == "off" {
			t.PreWarm = ""
		} else {
			t.PreWarm = val
		}
		d.Cfg.Targets[targetName] = t
		d.persistConfig()
	})
	preWarmSelect.Selected = currentPreWarm

	form := widget.NewForm(
		widget.NewFormItem("Auto-stop", autoStopSelect),
		widget.NewFormItem("Pre-warm", preWarmSelect),
	)

	return container.NewVBox(heading, form)
}

// persistConfig saves config and rebuilds the tray menu.
func (d *Daemon) persistConfig() {
	if err := config.SaveConfig(d.ConfigPath, d.Cfg); err != nil {
		log.Printf("error saving config: %v", err)
	}
	d.rebuildTrayMenu()
}
