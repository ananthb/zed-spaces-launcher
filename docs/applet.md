# Menu Bar Applet

The applet runs as a background process providing quick access to your codespaces from the system tray.

```bash
cosmonaut applet
```

## Features

- **System tray icon** — Z-in-cloud icon; hollow when idle, filled when tracking active codespaces
- **Global hotkey** — configurable shortcut (default `Cmd+Shift+S` on macOS, `Ctrl+Shift+S` on Linux)
- **Tray menu** — default target, recent repos, full picker
- **Codespace polling** — monitors running codespaces, sends desktop notifications on state changes
- **Auto-stop** — stops idle codespaces after a configured duration
- **Pre-warm** — creates/starts codespaces on a daily schedule before work hours

## Hotkey actions

The `hotkeyAction` config option controls what happens when you press the hotkey:

| Value | Behavior |
|---|---|
| `picker` (default) | Opens the interactive repo/codespace picker in a native window |
| `previous` | Launches the most recently used target from history |
| `default` | Launches the config's `defaultTarget` |

Both `previous` and `default` fall back to the picker if there's no history or no default target.

## Quick reconnect

When you click a repository in the tray menu, the applet passes the repo (or config target name) directly to the child process, skipping the repo selector TUI. If there is only one codespace for that repo and it is already running with an SSH config on disk, the applet focuses the existing Zed window without any intermediate UI.

## Service management

The applet is designed to be managed by your OS service manager.

### macOS (launchd)

The home-manager module automatically creates a launchd agent when `daemon.enable` is set (defaults to `true`).

### Linux (systemd)

The home-manager module creates a systemd user service, or use the included service file manually:

```bash
cp cosmonaut.service ~/.config/systemd/user/
systemctl --user enable --now cosmonaut
```

## Config

See the `daemon` section in [Configuration](config.md#daemon-fields).
