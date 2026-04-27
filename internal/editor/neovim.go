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
	return nil
}

func (n *NeovimEditor) LaunchRemote(sshAlias, workspacePath string) error {
	if runtime.GOOS == "darwin" {
		return launchInTerminalDarwin(n.Terminal, sshAlias, workspacePath)
	}
	return launchInTerminalLinux(n.Terminal, sshAlias, workspacePath)
}

func launchInTerminalDarwin(terminal, sshAlias, workspacePath string) error {
	sshCmd := fmt.Sprintf("ssh -t %s 'cd %s && nvim'", sshAlias, workspacePath)

	if terminal != "" && terminal != "auto" {
		// User specified a terminal app.
		script := fmt.Sprintf(
			`tell application %q to activate
tell application %q
do script %q
end tell`, terminal, terminal, sshCmd)
		return exec.Command("osascript", "-e", script).Run()
	}

	// Default: open in Terminal.app
	script := fmt.Sprintf(
		`tell application "Terminal"
activate
do script %q
end tell`, sshCmd)
	return exec.Command("osascript", "-e", script).Run()
}

func launchInTerminalLinux(terminal, sshAlias, workspacePath string) error {
	sshCmd := fmt.Sprintf("ssh -t %s 'cd %s && nvim'", sshAlias, workspacePath)

	if terminal != "" && terminal != "auto" {
		return exec.Command(terminal, "-e", "sh", "-c", sshCmd).Run()
	}
	for _, term := range []string{"ghostty", "alacritty", "kitty", "gnome-terminal", "xterm"} {
		if _, err := exec.LookPath(term); err == nil {
			return exec.Command(term, "-e", "sh", "-c", sshCmd).Run()
		}
	}
	return fmt.Errorf("no terminal emulator found; set daemon.terminal in config")
}
