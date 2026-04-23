# Usage

```
cosmonaut [target] [flags]
```

Run `cosmonaut --help` for flag and config field documentation.

## Interactive mode (no target)

When no target is given, the tool enters a TUI:

1. **Repository picker** — shows repos from your codespaces, config targets, and all your GitHub repositories. Type to filter, or type `owner/repo` to use any repository.
2. **Codespace selector** — pick an existing codespace or create a new one. Press `d` to delete a codespace, `esc` to choose a different repo.

If a `defaultTarget` is set in the config, the tool jumps straight to the codespace selector for that repo. Press `esc` to go back to the full repo picker.

## Auto-select

In non-interactive mode (e.g. when invoked from scripts or the applet), if a repository has exactly one codespace the selector is skipped. When the codespace is already `Available` and an SSH config already exists under `~/.ssh/cosmonaut/`, the tool also skips the SSH wait and config fetch, launching the editor immediately.

## Named target

```bash
cosmonaut work
```

Uses the target definition from the config file directly.

## Direct repository

```bash
cosmonaut owner/repo
```

When the argument contains a `/`, it is treated as a repository name rather than a config target. This builds a default target for that repo (workspace path `/workspaces/<repo>`) and skips the repo picker.

## Flags

| Flag | Description |
|---|---|
| `--config <path>` | Config file path (default `cosmonaut.config.json`) |
| `--codespace <name>` | Launch a specific codespace by name, skipping the selector |
| `--editor <name>` | Editor to launch: `zed` (default) or `neovim` |
| `--no-open` | Update SSH/editor config and print the `ssh://` target without launching the editor |
| `--dry-run` | Do not create a codespace or launch the editor |

The editor can also be set per-config via the top-level `editor` field.

## What it updates

| Path | Description |
|---|---|
| `~/.ssh/config` | Ensures `Include ~/.ssh/cosmonaut/*.conf` exists |
| `~/.ssh/cosmonaut/<codespace>.conf` | OpenSSH config from `gh codespace ssh --config` |
| Zed's `settings.json` | Upserts one `ssh_connections` entry per host alias (Zed editor only) |

Launching with `--editor neovim` opens an SSH session in a terminal emulator and runs `nvim`; it does not write to Zed's settings.

## Shell completions

Completions are installed automatically by the nix package. To set them up manually:

=== "bash"

    ```bash
    cosmonaut completion bash > /etc/bash_completion.d/cosmonaut
    ```

=== "zsh"

    ```bash
    cosmonaut completion zsh > "${fpath[1]}/_cosmonaut"
    ```

=== "fish"

    ```bash
    cosmonaut completion fish > ~/.config/fish/completions/cosmonaut.fish
    ```
