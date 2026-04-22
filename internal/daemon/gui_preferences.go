package daemon

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/ananth/cosmonaut/internal/codespace"
	"github.com/ananth/cosmonaut/internal/config"
)

// showPreferences opens a preferences window.
func (d *Daemon) showPreferences() {
	if d.app == nil {
		return
	}

	fyne.Do(func() {
		win := d.app.NewWindow("Preferences")
		win.Resize(fyne.NewSize(420, 380))
		win.SetFixedSize(true)
		win.CenterOnScreen()
		win.SetContent(d.buildPreferencesContent(win))
		win.Show()
	})
}

func (d *Daemon) buildPreferencesContent(win fyne.Window) fyne.CanvasObject {
	var items []fyne.CanvasObject

	// GitHub auth section.
	items = append(items, d.buildAuthSection(win))
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

	return container.NewPadded(container.NewVBox(items...))
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

	var actionBtn *widget.Button
	if authed {
		actionBtn = widget.NewButton("Remove auth", func() {
			go func() {
				_, _ = d.Runner.Run([]string{"auth", "logout", "--hostname", "github.com", "--yes"})
				fyne.Do(func() {
					win.SetContent(d.buildPreferencesContent(win))
					d.rebuildTrayMenu()
				})
			}()
		})
	} else {
		actionBtn = widget.NewButton("Log in...", func() {
			go func() {
				// Run gh auth login --web; it opens the browser automatically.
				_, err := d.Runner.Run([]string{"auth", "login", "--web", "--hostname", "github.com"})
				if err != nil {
					log.Printf("auth login: %v", err)
				}
				fyne.Do(func() {
					win.SetContent(d.buildPreferencesContent(win))
					d.rebuildTrayMenu()
				})
			}()
		})
	}

	return container.NewHBox(statusLabel, actionBtn)
}

func (d *Daemon) buildDaemonSection() fyne.CanvasObject {
	daemon := d.Cfg.Daemon

	// Hotkey action.
	currentAction := daemon.HotkeyAction
	if currentAction == "" {
		currentAction = "picker"
	}
	actionSelect := widget.NewSelect([]string{"picker", "previous", "default"}, func(val string) {
		d.Cfg.Daemon.HotkeyAction = val
		d.persistConfig()
	})
	actionSelect.Selected = currentAction

	// Poll interval.
	currentPoll := daemon.PollInterval
	if currentPoll == "" {
		currentPoll = "5m"
	}
	pollSelect := widget.NewSelect([]string{"1m", "5m", "15m", "30m"}, func(val string) {
		d.Cfg.Daemon.PollInterval = val
		d.persistConfig()
	})
	pollSelect.Selected = currentPoll

	form := widget.NewForm(
		widget.NewFormItem("Hotkey action", actionSelect),
		widget.NewFormItem("Poll interval", pollSelect),
	)
	return form
}

func (d *Daemon) buildTargetSection() fyne.CanvasObject {
	targetName := d.Cfg.DefaultTarget
	t := d.Cfg.Targets[targetName]

	heading := widget.NewLabel(fmt.Sprintf("Target: %s", targetName))
	heading.TextStyle = fyne.TextStyle{Bold: true}

	// Auto-stop.
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

	// Pre-warm.
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
