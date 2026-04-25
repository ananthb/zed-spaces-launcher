# cosmonaut

<p align="center">
  <img src="logo.svg" width="128" height="128" alt="cosmonaut logo">
</p>

CLI and menu bar applet for starting or creating GitHub Codespaces and opening them in [Zed](https://zed.dev) or Neovim via SSH remoting.

## What it does

1. Authenticates with GitHub via the `gh` CLI
2. Lets you pick a repository and codespace (interactive TUI with type-ahead filtering)
3. Creates a codespace if none exists
4. Writes an SSH config entry for the codespace and updates Zed's `settings.json` when launching Zed
5. Launches the editor with the remote workspace

## Quick start

```bash
# Install with Nix
nix profile install github:linuskendall/cosmonaut

# Or build from a git checkout
git clone https://github.com/linuskendall/cosmonaut
cd cosmonaut && go build -o cosmonaut .

# Run interactively
cosmonaut

# Or with a named target from config
cosmonaut work
```

## Menu bar applet

The applet provides a system tray icon with quick access to your codespaces, a global hotkey, scheduled pre-warm, and an optional sleep inhibitor while SSH sessions are active.

```bash
cosmonaut applet
```

See [Menu Bar Applet](applet.md) for details.
