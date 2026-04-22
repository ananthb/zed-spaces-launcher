package editor

import (
	"fmt"
	"os/exec"
	"runtime"
)

// NeovimEditor implements Editor for Neovim via SSH.
type NeovimEditor struct {
	// Terminal is the terminal app to use for launching Neovim.
	// Empty string means auto-detect.
	Terminal string
}

func (n *NeovimEditor) Name() string { return "neovim" }

func (n *NeovimEditor) FindBinary() (string, error) {
	if p, err := exec.LookPath("nvim"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("Neovim not found on PATH (tried \"nvim\")")
}

func (n *NeovimEditor) ConfigureConnection(sshAlias, workspacePath, nickname string, uploadBinary *bool) error {
	// Neovim has no equivalent of Zed's settings.json SSH connections.
	return nil
}

func (n *NeovimEditor) LaunchRemote(sshAlias, workspacePath string) error {
	// Launch SSH into the codespace with nvim as the command.
	sshCmd := fmt.Sprintf("cd %s && nvim", workspacePath)

	if runtime.GOOS == "darwin" {
		return launchInTerminalDarwin(n.Terminal, sshAlias, sshCmd)
	}
	return launchInTerminalLinux(n.Terminal, sshAlias, sshCmd)
}

func launchInTerminalDarwin(terminal, sshAlias, sshCmd string) error {
	if terminal == "" || terminal == "auto" {
		// Try common terminals in order of preference.
		for _, app := range []string{"iTerm", "Ghostty", "Alacritty", "Terminal"} {
			if err := exec.Command("open", "-a", app,
				"--args", "ssh", "-t", sshAlias, sshCmd).Run(); err == nil {
				return nil
			}
		}
		// Fallback: use osascript to open Terminal.app with ssh command.
		script := fmt.Sprintf(
			`tell application "Terminal" to do script "ssh -t %s '%s'"`,
			sshAlias, sshCmd,
		)
		return exec.Command("osascript", "-e", script).Run()
	}
	return exec.Command("open", "-a", terminal,
		"--args", "ssh", "-t", sshAlias, sshCmd).Run()
}

func launchInTerminalLinux(terminal, sshAlias, sshCmd string) error {
	if terminal == "" || terminal == "auto" {
		// Try common terminals.
		for _, term := range []string{"ghostty", "alacritty", "kitty", "gnome-terminal", "xterm"} {
			if _, err := exec.LookPath(term); err == nil {
				return exec.Command(term, "-e", "ssh", "-t", sshAlias, sshCmd).Run()
			}
		}
		return fmt.Errorf("no terminal emulator found; set daemon.terminal in config")
	}
	return exec.Command(terminal, "-e", "ssh", "-t", sshAlias, sshCmd).Run()
}
