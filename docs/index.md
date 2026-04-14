# codespace-zed

<p align="center">
  <img src="logo.svg" width="128" height="128" alt="codespace-zed logo">
</p>

CLI and menu bar applet for starting or creating GitHub Codespaces and opening them in [Zed](https://zed.dev) via SSH remoting.

## What it does

1. Authenticates with GitHub via the `gh` CLI
2. Lets you pick a repository and codespace (interactive TUI with type-ahead filtering)
3. Creates a codespace if none exists
4. Configures SSH and Zed's remote connection settings
5. Launches Zed with the remote workspace

## Quick start

```bash
# Install with Nix
nix profile install github:linuskendall/zed-spaces-launcher

# Or build from source
go install github.com/ananth/codespace-zed@latest

# Run interactively
codespace-zed

# Or with a named target from config
codespace-zed work
```

## Menu bar applet

The applet provides a system tray icon with quick access to your codespaces, a global hotkey, and lifecycle management (auto-stop, pre-warm).

```bash
codespace-zed applet
```

See [Menu Bar Applet](applet.md) for details.
