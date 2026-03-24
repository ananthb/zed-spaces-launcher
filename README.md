# Codespace Zed Launcher

Local CLI for starting or creating a predefined GitHub Codespace, wiring it into Zed's SSH remoting config, and opening the remote workspace in Zed.

## Requirements

- `python3`
- `gh` installed and authenticated with `gh auth login`
- `zed` installed locally
- GitHub Codespaces image includes an SSH server
  - For standard dev containers, GitHub documents the `ghcr.io/devcontainers/features/sshd:1` feature for this

## Config

Copy `codespace-zed.config.example.json` to `codespace-zed.config.json` and define one or more named targets:

- `repository`: GitHub repository in `owner/repo` form
- `branch`: preferred branch when creating or matching a codespace
- `displayName`: optional exact display name to disambiguate matches
- `codespaceName`: optional exact codespace name if you want strict reuse
- `workspacePath`: remote folder Zed should open
- `machine`, `location`, `devcontainerPath`, `idleTimeout`, `retentionPeriod`: forwarded to `gh codespace create`
- `uploadBinaryOverSsh`: writes Zed's `upload_binary_over_ssh` setting for that host
- `zedNickname`: friendly name shown in Zed's remote project list

## Usage

```bash
cp codespace-zed.config.example.json codespace-zed.config.json
python3 ./bin/codespace-zed demo
```

Flags:

- `--config <path>`: use a non-default config file
- `--no-open`: update SSH/Zed config and print the final `ssh://` target without launching Zed
- `--dry-run`: do not create a codespace or launch Zed

## What it updates

- `~/.ssh/config`
  - Ensures `Include ~/.ssh/codespaces-zed/*.conf` exists
- `~/.ssh/codespaces-zed/<codespace>.conf`
  - Stores the OpenSSH config emitted by `gh codespace ssh --config`
- `~/.zed/settings.json` on macOS, or `~/.config/zed/settings.json` on Linux
  - Upserts one `ssh_connections` entry per generated host alias

## Validation

```bash
python3 -m unittest
python3 -m py_compile codespace_zed.py bin/codespace-zed
python3 ./bin/codespace-zed demo --dry-run
```

The final dry run still needs live GitHub API access because `gh codespace list` is the source of truth for whether an existing codespace already matches the target.
