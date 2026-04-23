# Installation

## Requirements

- [`gh`](https://cli.github.com/) installed and authenticated (`gh auth login`)
- [`zed`](https://zed.dev) (default) or `nvim` installed locally, matching the editor you plan to launch
- GitHub Codespaces image includes an SSH server
    - For standard dev containers, use the [`ghcr.io/devcontainers/features/sshd:1`](https://github.com/devcontainers/features/tree/main/src/sshd) feature

## macOS

Download the `.dmg` from [GitHub Releases](https://github.com/linuskendall/zed-spaces-launcher/releases) and open it. The DMG contains the `cosmonaut` binary and an example config. Copy the binary somewhere on `PATH`:

```bash
sudo cp /Volumes/cosmonaut/cosmonaut /usr/local/bin/
xattr -d com.apple.quarantine /usr/local/bin/cosmonaut
```

The `xattr` step clears the Gatekeeper quarantine flag that macOS applies to files pulled from a downloaded DMG.

Available for: `aarch64` (Apple Silicon).

## Linux

### AppImage

Download the `.AppImage` from [GitHub Releases](https://github.com/linuskendall/zed-spaces-launcher/releases):

```bash
chmod +x cosmonaut-*.AppImage
./cosmonaut-*.AppImage
```

Available for: `amd64`.

### Tarball

Download the `.tar.gz` from [GitHub Releases](https://github.com/linuskendall/zed-spaces-launcher/releases). Each tarball includes the binary, an example config, and a systemd user service file.

```bash
tar xzf cosmonaut-amd64.tar.gz
sudo cp cosmonaut/cosmonaut /usr/local/bin/
# Optional: install systemd user service
cp cosmonaut/cosmonaut.service ~/.config/systemd/user/
systemctl --user enable --now cosmonaut
```

Available for: `amd64`.

## Nix flake

```nix
# flake.nix
{
  inputs.cosmonaut.url = "github:linuskendall/zed-spaces-launcher";
}
```

The package includes shell completions for bash, zsh, and fish.

## Home Manager

```nix
{
  imports = [ cosmonaut.homeManagerModules.default ];

  programs.cosmonaut = {
    enable = true;
    defaultTarget = "work";
    targets.work = {
      repository = "my-org/my-repo";
      branch = "main";
      machine = "standardLinux32gb";
    };
  };
}
```

This generates the config file, wraps the binary with `--config`, sets up SSH includes, installs shell completions, and runs the menu bar applet via launchd (macOS) or systemd (Linux).

## From source

```bash
git clone https://github.com/linuskendall/zed-spaces-launcher
cd zed-spaces-launcher
go build -o cosmonaut .
```

Requires Go 1.26+ and CGo (for the Fyne GUI toolkit used by the applet).

## Verifying releases

All release artifacts are signed with [Sigstore cosign](https://docs.sigstore.dev/) (keyless). Each file has a corresponding `.sig` (signature) and `.pem` (certificate).

### Verify a file

```bash
cosign verify-blob \
  --certificate cosmonaut-amd64.tar.gz.pem \
  --signature cosmonaut-amd64.tar.gz.sig \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp 'github.com/linuskendall/zed-spaces-launcher' \
  cosmonaut-amd64.tar.gz
```

### Verify checksums

```bash
# Verify the SHA256SUMS file itself was signed by the release workflow
cosign verify-blob \
  --certificate SHA256SUMS.pem \
  --signature SHA256SUMS.sig \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp 'github.com/linuskendall/zed-spaces-launcher' \
  SHA256SUMS

# Then verify file integrity
sha256sum -c SHA256SUMS
```
