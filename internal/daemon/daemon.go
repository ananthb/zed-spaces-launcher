// Package daemon implements the cosmonaut menu bar applet which
// hosts a system tray icon, global hotkey listener, and codespace
// lifecycle management.
package daemon

import (
	"log"
	"os"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"

	"github.com/ananth/cosmonaut/internal/codespace"
	"github.com/ananth/cosmonaut/internal/config"
)

// Daemon is the long-running background process that hosts the system tray,
// hotkey listener, and codespace poller.
type Daemon struct {
	Cfg        *config.Config
	ConfigPath string
	Runner     codespace.GHRunner

	app fyne.App

	mu         sync.Mutex
	codespaces []codespace.Codespace
	stopCh     chan struct{}
}

// New creates a new Daemon with the given config.
func New(cfg *config.Config, configPath string) *Daemon {
	return &Daemon{
		Cfg:        cfg,
		ConfigPath: configPath,
		Runner:     codespace.DefaultGHRunner{},
		stopCh:     make(chan struct{}),
	}
}

// Run starts all applet components. It blocks until Stop is called.
// This must be called from the main OS thread.
func (d *Daemon) Run() error {
	enrichPath()

	d.app = app.NewWithID("dev.cosmonaut.applet")
	d.app.Settings().SetTheme(newCosmoTheme())
	d.app.SetIcon(appIcon())

	log.Printf("applet started (pid %d)", os.Getpid())

	// Run the initial poll synchronously so the tray menu has
	// codespace data before it is first displayed.
	d.poll()

	// Start background workers.
	go d.startPoller()
	go d.startHotkeyListener()
	d.startPreWarm()

	// Create a hidden master window so popover windows don't quit the app on close.
	master := d.app.NewWindow("cosmonaut-applet")
	master.SetMaster()
	master.SetCloseIntercept(func() {})

	// Set up system tray.
	if desk, ok := d.app.(desktop.App); ok {
		desk.SetSystemTrayIcon(trayIconIdle())
		desk.SetSystemTrayMenu(d.buildTrayMenu())
	}

	// Run the Fyne event loop (blocks until Quit).
	d.app.Run()

	return nil
}

// Stop signals the applet to shut down.
func (d *Daemon) Stop() {
	select {
	case <-d.stopCh:
		return
	default:
		close(d.stopCh)
	}

	if d.app != nil {
		d.app.Quit()
	}

	log.Println("applet stopped")
}

// Codespaces returns the last-polled codespace list.
func (d *Daemon) Codespaces() []codespace.Codespace {
	d.mu.Lock()
	defer d.mu.Unlock()
	result := make([]codespace.Codespace, len(d.codespaces))
	copy(result, d.codespaces)
	return result
}

// SetCodespaces updates the cached codespace list.
func (d *Daemon) SetCodespaces(cs []codespace.Codespace) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.codespaces = cs
}
