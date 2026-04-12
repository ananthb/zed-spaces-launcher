# Usage

```
codespace-zed [target] [flags]
```

Run `codespace-zed --help` for flag and config field documentation.

## Interactive mode (no target)

When no target is given, the tool enters a TUI:

1. **Repository picker** — shows repos from your codespaces, config targets, and all your GitHub repositories. Type to filter, or type `owner/repo` to use any repository.
2. **Codespace selector** — pick an existing codespace or create a new one. Press `d` to delete a codespace, `esc` to choose a different repo.

If a `defaultTarget` is set in the config, the tool jumps straight to the codespace selector for that repo. Press `esc` to go back to the full repo picker.

## Named target

```bash
codespace-zed work
```

Uses the target definition from the config file directly.

## Flags

| Flag | Description |
|---|---|
| `--config <path>` | Config file path (default `codespace-zed.config.json`) |
| `--no-open` | Update SSH/Zed config and print the `ssh://` target without launching Zed |
| `--dry-run` | Do not create a codespace or launch Zed |

## What it updates

| Path | Description |
|---|---|
| `~/.ssh/config` | Ensures `Include ~/.ssh/codespaces-zed/*.conf` exists |
| `~/.ssh/codespaces-zed/<codespace>.conf` | OpenSSH config from `gh codespace ssh --config` |
| Zed's `settings.json` | Upserts one `ssh_connections` entry per host alias |

## Shell completions

Completions are installed automatically by the nix package. To set them up manually:

=== "bash"

    ```bash
    codespace-zed completion bash > /etc/bash_completion.d/codespace-zed
    ```

=== "zsh"

    ```bash
    codespace-zed completion zsh > "${fpath[1]}/_codespace-zed"
    ```

=== "fish"

    ```bash
    codespace-zed completion fish > ~/.config/fish/completions/codespace-zed.fish
    ```
