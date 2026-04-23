# Configuration

cosmonaut uses a JSONC config file (comments and trailing commas are supported).

The default path is `cosmonaut.config.json` in the current directory for the CLI, or `~/.config/cosmonaut/config.json` (XDG) for the applet.

## Example

```json
{
  // Default target to use when no target name is given.
  "defaultTarget": "work",

  // Applet settings (menu bar icon, hotkey, codespace lifecycle).
  "daemon": {
    "hotkey": "Cmd+Shift+S",
    "hotkeyAction": "picker",
    "terminal": "auto",
    "pollInterval": "5m"
  },

  "targets": {
    "work": {
      "repository": "my-org/my-repo",
      "branch": "main",
      "workspacePath": "/workspaces/my-repo",
      "machine": "standardLinux32gb",
      "autoStop": "30m",
      "preWarm": "08:00"
    }
  }
}
```

## Target fields

| Field | Type | Required | Description |
|---|---|---|---|
| `repository` | string | yes | GitHub repository in `owner/repo` form |
| `branch` | string | | Preferred branch when creating or matching a codespace |
| `displayName` | string | | Exact display name to disambiguate codespace matches |
| `codespaceName` | string | | Exact codespace name for strict reuse |
| `workspacePath` | string | yes | Remote folder Zed should open (e.g. `/workspaces/repo`) |
| `machine` | string | | Machine type forwarded to `gh codespace create` |
| `location` | string | | Location forwarded to `gh codespace create` |
| `devcontainerPath` | string | | Dev container config path |
| `idleTimeout` | string | | Idle timeout (e.g. `30m`) |
| `retentionPeriod` | string | | Retention period (e.g. `720h`) |
| `uploadBinaryOverSsh` | bool | | Zed's `upload_binary_over_ssh` setting |
| `zedNickname` | string | | Friendly name in Zed's remote project list |
| `autoStop` | string | | Auto-stop after idle duration (applet only, e.g. `30m`) |
| `preWarm` | string | | Time-of-day to pre-warm codespace (applet only, e.g. `08:00`) |

## Daemon fields

These go in the top-level `"daemon"` object and configure the menu bar applet.

| Field | Type | Description |
|---|---|---|
| `hotkey` | string | Global hotkey (e.g. `Cmd+Shift+S` or `Ctrl+Shift+S`) |
| `hotkeyAction` | string | `picker` (default), `previous`, or `default` |
| `terminal` | string | Terminal app for picker; `auto` to detect |
| `pollInterval` | string | Codespace poll interval (e.g. `5m`) |
| `inhibitSleep` | string | Hold a sleep inhibitor while a launched SSH session is alive: `off` (default), `sleep`, or `sleep+shutdown`. On macOS, `sleep+shutdown` degrades to `sleep` (no user-space shutdown inhibitor). |
