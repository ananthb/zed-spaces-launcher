<p align="center">
  <img src="docs/logo.svg" width="128" height="128" alt="codespace-zed logo">
</p>

# Codespace Zed Launcher

CLI and menu bar applet for GitHub Codespaces + [Zed](https://zed.dev).

**[Documentation](https://ananthb.github.io/zed-spaces-launcher/)** ·
**[Configuration](https://ananthb.github.io/zed-spaces-launcher/config/)** ·
**[API Reference](https://ananthb.github.io/zed-spaces-launcher/api/)**

## Install

### macOS

Download the `.dmg` from [Releases](https://github.com/ananthb/zed-spaces-launcher/releases), open it, and drag to Applications.

### Linux

Download the `.AppImage` from [Releases](https://github.com/ananthb/zed-spaces-launcher/releases):

```bash
chmod +x codespace-zed-*.AppImage
./codespace-zed-*.AppImage
```

### Nix

```nix
{
  inputs.codespace-zed.url = "github:ananthb/zed-spaces-launcher";
}
```

Or with [Home Manager](https://ananthb.github.io/zed-spaces-launcher/install/#home-manager) for declarative config + auto-start.

## Quick start

```bash
# Launch interactively — pick a repo, select a codespace, open in Zed
codespace-zed

# Or use a named target from your config
codespace-zed work

# Start the menu bar applet (tray icon, hotkey, lifecycle management)
codespace-zed applet
```

## Requirements

- [`gh`](https://cli.github.com/) authenticated (`gh auth login`)
- [`zed`](https://zed.dev) installed
- SSH server in your codespace image ([`ghcr.io/devcontainers/features/sshd:1`](https://github.com/devcontainers/features/tree/main/src/sshd))
