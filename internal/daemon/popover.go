package daemon

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"fyne.io/fyne/v2"
	"github.com/creack/pty"
	terminal "github.com/fyne-io/terminal"
)

// findBinary locates the codespace-zed wrapper binary to spawn in the popover.
// Order: PATH lookup → directory of os.Executable() → os.Executable() itself.
func findBinary() (string, error) {
	// Best case: on PATH (works in dev, may not work under launchd).
	if p, err := exec.LookPath("codespace-zed"); err == nil {
		return p, nil
	}

	// Under launchd/systemd the service runs the inner binary directly.
	// The outermost wrapper (with gh on PATH + --config) lives in the
	// same bin/ directory as os.Executable(), named "codespace-zed".
	self, err := os.Executable()
	if err != nil {
		return "", err
	}
	wrapper := filepath.Join(filepath.Dir(self), "codespace-zed")
	if _, err := os.Stat(wrapper); err == nil {
		return wrapper, nil
	}

	// Last resort: use ourselves.
	return self, nil
}

// showPopover opens a Fyne window with a terminal widget
// running codespace-zed with the given args.
func (d *Daemon) showPopover(args ...string) {
	if d.app == nil {
		log.Println("popover: app not initialized")
		return
	}

	binary, err := findBinary()
	if err != nil {
		log.Printf("popover: cannot find binary: %v", err)
		return
	}

	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	log.Printf("popover: launching %s %v", binary, args)

	// Start with an initial PTY size.
	initialSize := &pty.Winsize{Rows: 30, Cols: 80}
	ptmx, err := pty.StartWithSize(cmd, initialSize)
	if err != nil {
		log.Printf("popover: pty start: %v", err)
		return
	}

	log.Printf("popover: child pid %d", cmd.Process.Pid)

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

	// Wait for child in background so we can log exit status.
	childDone := make(chan error, 1)
	go func() {
		childDone <- cmd.Wait()
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

	// Run the terminal connected to the PTY (blocks until EOF on reader).
	if err := term.RunWithConnection(ptmx, ptmx); err != nil && err != io.EOF {
		log.Printf("popover: terminal: %v", err)
	}

	// Check why the child exited.
	select {
	case err := <-childDone:
		if err != nil {
			log.Printf("popover: child exited with error: %v", err)
		} else {
			log.Printf("popover: child exited cleanly")
		}
	default:
		log.Printf("popover: terminal disconnected, cleaning up")
	}

	// Process exited — clean up and close the window.
	cleanup()
	fyne.Do(func() {
		if w != nil {
			w.Close()
		}
	})
}
