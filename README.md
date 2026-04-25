<p align="center">
  <img src="docs/logo.svg" width="128" height="128" alt="cosmonaut logo">
</p>

# Cosmonaut Launcher

CLI and menu bar applet for GitHub Codespaces + [Zed](https://zed.dev).

**[Documentation](https://linuskendall.github.io/cosmonaut/)** ·
**[Configuration](https://linuskendall.github.io/cosmonaut/config/)** ·
**[API Reference](https://linuskendall.github.io/cosmonaut/api/)**

## Install

### macOS

Download the `.dmg` from [Releases](https://github.com/linuskendall/cosmonaut/releases), open it, and drag to Applications.

### Linux

Download the `.AppImage` from [Releases](https://github.com/linuskendall/cosmonaut/releases):

```bash
chmod +x cosmonaut-*.AppImage
./cosmonaut-*.AppImage
```

### Nix

```nix
{
  inputs.cosmonaut.url = "github:linuskendall/cosmonaut";
}
```

Or with [Home Manager](https://linuskendall.github.io/cosmonaut/install/#home-manager) for declarative config + auto-start.

## Quick start

```bash
# Launch interactively: pick a repo, select a codespace, open in Zed
cosmonaut

# Or use a named target from your config
cosmonaut work

# Start the menu bar applet (tray icon, hotkey, lifecycle management)
cosmonaut applet
```

## Requirements

- [`gh`](https://cli.github.com/) authenticated (`gh auth login`)
- [`zed`](https://zed.dev) installed
- SSH server in your codespace image ([`ghcr.io/devcontainers/features/sshd:1`](https://github.com/devcontainers/features/tree/main/src/sshd))
