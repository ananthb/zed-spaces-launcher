package daemon

import (
	"io"
	"log"
	"os"
	"os/exec"
	"sync"

	"fyne.io/fyne/v2"
	"github.com/creack/pty"
	terminal "github.com/fyne-io/terminal"
)

// showPopover opens a Fyne window with a terminal widget
// running codespace-zed with the given args.
func (d *Daemon) showPopover(args ...string) {
	if d.app == nil {
		log.Println("popover: app not initialized")
		return
	}

	binary, err := os.Executable()
	if err != nil {
		log.Printf("popover: cannot find executable: %v", err)
		return
	}

	// Build the command. Pass --config so the child uses the same config.
	cmdArgs := []string{binary}
	if d.ConfigPath != "" {
		cmdArgs = append(cmdArgs, "--config", d.ConfigPath)
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	// Start with an initial PTY size.
	initialSize := &pty.Winsize{Rows: 30, Cols: 80}
	ptmx, err := pty.StartWithSize(cmd, initialSize)
	if err != nil {
		log.Printf("popover: pty start: %v", err)
		return
	}

	term := terminal.New()

	// done is closed when the popover is being torn down.
	done := make(chan struct{})
	var closeOnce sync.Once
	cleanup := func() {
		closeOnce.Do(func() {
			close(done)
			ptmx.Close()
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		})
	}

	// Listen for terminal resize events and update the PTY.
	configCh := make(chan terminal.Config, 1)
	term.AddListener(configCh)
	go func() {
		defer term.RemoveListener(configCh)
		for {
			select {
			case cfg, ok := <-configCh:
				if !ok {
					return
				}
				if cfg.Rows > 0 && cfg.Columns > 0 {
					pty.Setsize(ptmx, &pty.Winsize{
						Rows: uint16(cfg.Rows),
						Cols: uint16(cfg.Columns),
					})
				}
			case <-done:
				return
			}
		}
	}()

	var w fyne.Window
	fyne.Do(func() {
		w = d.app.NewWindow("codespace-zed")
		w.SetPadded(false)
		w.Resize(fyne.NewSize(700, 500))
		w.CenterOnScreen()
		w.SetContent(term)
		w.SetOnClosed(func() {
			cleanup()
		})
		w.Show()
	})

	// Run the terminal connected to the PTY (blocks until process exits).
	if err := term.RunWithConnection(ptmx, ptmx); err != nil && err != io.EOF {
		log.Printf("popover: terminal: %v", err)
	}

	// Process exited — clean up and close the window.
	cleanup()
	fyne.Do(func() {
		if w != nil {
			w.Close()
		}
	})
}
