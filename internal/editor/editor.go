// Package editor abstracts editor-specific operations for launching and
// configuring remote connections to codespaces. Implementations exist for
// Zed and Neovim.
package editor

import "fmt"

// Editor abstracts editor-specific operations.
type Editor interface {
	// Name returns the editor identifier (e.g. "zed", "neovim").
	Name() string
	// FindBinary locates the editor's CLI binary on PATH.
	FindBinary() (string, error)
	// ConfigureConnection sets up editor-specific config for the SSH
	// connection (e.g. Zed's settings.json). No-op for editors that
	// don't need it.
	ConfigureConnection(sshAlias, workspacePath, nickname string, uploadBinary *bool) error
	// LaunchRemote opens the editor connected to the remote codespace.
	LaunchRemote(sshAlias, workspacePath string) error
}

// ForName returns an Editor implementation for the given name.
// An empty name defaults to "zed".
func ForName(name string) (Editor, error) {
	switch name {
	case "", "zed":
		return &ZedEditor{}, nil
	case "neovim", "nvim":
		return &NeovimEditor{}, nil
	default:
		return nil, fmt.Errorf("unknown editor %q (supported: zed, neovim)", name)
	}
}

// ResolveNickname determines the nickname for a connection.
// Checks zedNickname, targetDisplayName, codespaceDisplayName, then targetName.
func ResolveNickname(zedNickname, targetDisplayName, codespaceDisplayName, targetName string) string {
	if zedNickname != "" {
		return zedNickname
	}
	if targetDisplayName != "" {
		return targetDisplayName
	}
	if codespaceDisplayName != "" {
		return codespaceDisplayName
	}
	return targetName
}
