# Contributing

## Development setup

The project uses [Nix](https://nixos.org/) for builds and a dev shell.

```bash
# Enter dev shell with Go, gh, and other tools
nix develop

# Build the binary
nix build

# Run directly
go run .
```

## Testing

```bash
go test ./...
```

## Version bumps

The version number is specified in **three** places. All must be updated together:

| File | Field | Example |
|---|---|---|
| `flake.nix` | `version` | `version = "0.6.0";` |
| `dist/Info.plist` | `CFBundleShortVersionString` | `<string>0.6.0</string>` |
| `FyneApp.toml` | `Version` | `Version = "0.6.0"` |

The Nix Home Manager module in `modules/home-manager.nix` also duplicates some daemon config defaults (like the hotkey) for documentation purposes. If you change a default in the Go code, update the Nix module default to match.

## Releasing a new version

1. Bump the version in all three files listed above.
2. Commit and push to `main`.
3. Create an annotated tag and push it:
   ```bash
   git tag -a v0.7.0 -m "v0.7.0"
   git push origin v0.7.0
   ```
4. The [release workflow](https://github.com/linuskendall/zed-spaces-launcher/blob/main/.github/workflows/release.yml) runs automatically on tag push. It builds Linux (tarball + AppImage) and macOS (DMG) artifacts, generates checksums, signs them with cosign, and creates a GitHub release.

## Updating documentation

Docs live in `docs/` and are built with [MkDocs Material](https://squidfunk.github.io/mkdocs-material/). The site is configured in `mkdocs.yml`.

To preview locally:

```bash
mkdocs serve
```

API reference docs under `docs/api/` are generated from godoc comments using [gomarkdoc](https://github.com/princjef/gomarkdoc). The [docs workflow](https://github.com/linuskendall/zed-spaces-launcher/blob/main/.github/workflows/docs.yml) regenerates and deploys them on push to `main`.

To add a new page, create a markdown file in `docs/` and add it to the `nav` section in `mkdocs.yml`.
