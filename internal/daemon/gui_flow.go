package daemon

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/ananth/cosmonaut/internal/codespace"
	"github.com/ananth/cosmonaut/internal/config"
	"github.com/ananth/cosmonaut/internal/history"
	"github.com/ananth/cosmonaut/internal/slug"
	"github.com/ananth/cosmonaut/internal/sshconfig"
	"github.com/ananth/cosmonaut/internal/zed"
)

// showGUI opens the native GUI window. It replaces showPopover.
// Args are parsed the same way showPopover parsed them:
//   - no args: open repo picker
//   - target name or owner/repo: open codespace selector for that target
//   - "--codespace", csName, target: direct launch
func (d *Daemon) showGUI(args ...string) {
	if d.app == nil {
		log.Println("gui: app not initialized")
		return
	}

	// Parse args.
	var targetArg, codespaceName string
	for i := 0; i < len(args); i++ {
		if args[i] == "--codespace" && i+1 < len(args) {
			codespaceName = args[i+1]
			i++
		} else {
			targetArg = args[i]
		}
	}

	fyne.Do(func() {
		win := d.createGUIWindow("Cosmonaut")

		if codespaceName != "" && targetArg != "" {
			// Direct codespace launch.
			target, resolvedName := d.resolveTarget(targetArg)
			cs := &codespace.Codespace{Name: codespaceName, Repository: codespace.RepoField(target.Repository)}
			win.Show()
			d.runLaunchFlow(win, target, resolvedName, cs)
		} else if targetArg != "" {
			// Open codespace selector for a specific target/repo.
			target, resolvedName := d.resolveTarget(targetArg)
			d.showCodespaceSelector(win, target.Repository, target, resolvedName)
			win.Show()
		} else {
			// Open repo picker.
			d.showRepoPicker(win)
			win.Show()
		}
	})
}

// resolveTarget resolves a target argument to a config Target.
func (d *Daemon) resolveTarget(arg string) (config.Target, string) {
	if strings.Contains(arg, "/") {
		return guiTargetForRepo(d.Cfg, arg)
	}
	if d.Cfg != nil {
		if t, ok := d.Cfg.Targets[arg]; ok {
			return t, arg
		}
	}
	return guiTargetForRepo(d.Cfg, arg)
}

// showRepoPicker displays the repo picker screen.
func (d *Daemon) showRepoPicker(win fyne.Window) {
	screen := d.newRepoPickerScreen(win,
		func(repo string) {
			// Repo selected → show codespace selector.
			target, resolvedName := guiTargetForRepo(d.Cfg, repo)
			d.showCodespaceSelector(win, repo, target, resolvedName)
		},
		func() {
			win.Close()
		},
	)
	win.SetContent(screen.canvas)
}

// showCodespaceSelector displays the codespace selector screen.
func (d *Daemon) showCodespaceSelector(win fyne.Window, repo string, target config.Target, resolvedName string) {
	screen := d.newCodespaceScreen(win, repo, target, resolvedName,
		func(cs *codespace.Codespace) {
			if cs != nil {
				// Existing codespace selected → launch.
				d.runLaunchFlow(win, target, resolvedName, cs)
			} else {
				// Create new → show work label input.
				d.showWorkLabelInput(win, target, resolvedName)
			}
		},
		func() {
			// Back → repo picker.
			d.showRepoPicker(win)
		},
		func() {
			win.Close()
		},
	)
	win.SetContent(screen.canvas)
}

// showWorkLabelInput displays the work label input screen.
func (d *Daemon) showWorkLabelInput(win fyne.Window, target config.Target, resolvedName string) {
	screen := newWorkLabelScreen(
		func(label string) {
			// Create codespace with label.
			createTarget := target
			createTarget.DisplayName = slug.BuildDisplayName(
				target.Repository,
				target.Branch,
				label,
				target.DisplayName,
			)
			d.runCreateAndLaunch(win, createTarget, resolvedName)
		},
		func() {
			win.Close()
		},
	)
	win.SetContent(screen.canvas)
}

// runCreateAndLaunch creates a codespace and then launches it.
func (d *Daemon) runCreateAndLaunch(win fyne.Window, target config.Target, resolvedName string) {
	progress := newProgressScreen("Creating codespace...")
	win.SetContent(progress.canvas)

	go func() {
		cs, err := codespace.CreateCodespace(d.Runner, target)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("creating codespace: %w", err), win)
			})
			return
		}
		d.runLaunchFlow(win, target, resolvedName, cs)
	}()
}

// runLaunchFlow runs the SSH setup and Zed launch sequence with progress updates.
func (d *Daemon) runLaunchFlow(win fyne.Window, target config.Target, resolvedName string, selected *codespace.Codespace) {
	progress := newProgressScreen("Preparing codespace...")
	fyne.Do(func() {
		win.SetContent(progress.canvas)
	})

	go func() {
		setStatus := func(msg string) {
			fyne.Do(func() { progress.setStatus(msg) })
		}

		// Record in history.
		hist := history.Load()
		hist.Touch(target.Repository)
		hist.Save()

		// Fast path: if already Available with existing SSH config.
		if selected.State == "Available" {
			paths := sshconfig.ResolvePaths()
			if alias, ok := sshconfig.ReadExistingAlias(paths.IncludeDir, selected.Name); ok {
				remoteURL := fmt.Sprintf("ssh://%s/%s", alias, strings.TrimLeft(target.WorkspacePath, "/"))
				setStatus("Launching Zed...")
				if err := launchZed(remoteURL); err != nil {
					fyne.Do(func() { dialog.ShowError(err, win) })
					return
				}
				fyne.Do(func() { win.Close() })
				return
			}
		}

		// Ensure SSH connectivity.
		setStatus("Waiting for codespace SSH...")
		if err := codespace.EnsureReachable(d.Runner, selected.Name); err != nil {
			fyne.Do(func() { dialog.ShowError(fmt.Errorf("SSH connectivity: %w", err), win) })
			return
		}

		// Get SSH config.
		setStatus("Fetching SSH config...")
		sshCfg, err := codespace.GetSSHConfig(d.Runner, selected.Name)
		if err != nil {
			fyne.Do(func() { dialog.ShowError(err, win) })
			return
		}

		sshAlias, err := sshconfig.ParsePrimaryHostAlias(sshCfg)
		if err != nil {
			fyne.Do(func() { dialog.ShowError(err, win) })
			return
		}

		// Write SSH config.
		paths := sshconfig.ResolvePaths()
		if err := os.MkdirAll(paths.IncludeDir, 0700); err != nil {
			fyne.Do(func() { dialog.ShowError(err, win) })
			return
		}
		if err := sshconfig.EnsureConfigIncludesGenerated(paths.MainConfigPath); err != nil {
			fyne.Do(func() { dialog.ShowError(err, win) })
			return
		}
		if err := sshconfig.WriteCodespaceConfig(paths.IncludeDir, selected.Name, sshCfg); err != nil {
			fyne.Do(func() { dialog.ShowError(err, win) })
			return
		}

		// Update Zed settings.
		nickname := zed.ResolveNickname(
			target.ZedNickname,
			target.DisplayName,
			selected.DisplayName,
			resolvedName,
		)
		conn := zed.BuildConnection(sshAlias, target.WorkspacePath, nickname, target.UploadBinaryOverSSH)
		settingsPath := zed.ResolveSettingsPath()
		if err := zed.UpsertConnectionInFile(settingsPath, conn); err != nil {
			fyne.Do(func() { dialog.ShowError(err, win) })
			return
		}

		// Launch Zed.
		remoteURL := fmt.Sprintf("ssh://%s/%s", sshAlias, strings.TrimLeft(target.WorkspacePath, "/"))
		setStatus("Launching Zed...")
		if err := launchZed(remoteURL); err != nil {
			fyne.Do(func() { dialog.ShowError(err, win) })
			return
		}

		fyne.Do(func() { win.Close() })
		d.rebuildTrayMenu()
	}()
}

// launchZed finds the Zed binary and opens the remote URL.
func launchZed(remoteURL string) error {
	zedBin, err := codespace.FindZedBinary()
	if err != nil {
		return err
	}
	return exec.Command(zedBin, remoteURL).Run()
}
