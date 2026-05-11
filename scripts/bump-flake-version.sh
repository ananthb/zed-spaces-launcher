#!/usr/bin/env bash
# Update flake.nix's prebuilt-fetch package pinning after a release.
# Reads the latest release tag from `gh release view`, computes
# nix sha256 hashes for the user-consumable archives, and rewrites
# flake.nix.
#
# Lands meaningfully once PR 3 (prebuilt-fetch flake) is merged. Stub
# accepted in PR 1 so the script lives next to its peers.
#
# Usage: scripts/bump-flake-version.sh [tag]
#   tag defaults to the latest release.
set -euo pipefail

if ! command -v gh >/dev/null 2>&1; then
  echo "error: gh CLI not on PATH" >&2; exit 1
fi
if ! command -v nix-prefetch-url >/dev/null 2>&1; then
  echo "error: nix-prefetch-url not on PATH (run inside nix develop)" >&2; exit 1
fi

repo="${REPO:-linuskendall/cosmonaut}"
tag="${1:-$(gh release view --repo "$repo" --json tagName -q .tagName)}"
version="${tag#v}"

# Asset names are the canonical user-facing artifact names produced by
# goreleaser (see .goreleaser.{linux,darwin}.yaml). Update here if
# those name templates change.
linux_asset="cosmonaut-amd64.tar.gz"
darwin_asset="cosmonaut-macos-arm64.tar.gz"

echo "==> bumping flake.nix to $tag"

linux_url="https://github.com/$repo/releases/download/$tag/$linux_asset"
darwin_url="https://github.com/$repo/releases/download/$tag/$darwin_asset"

linux_sha=$(nix-prefetch-url --type sha256 "$linux_url")
darwin_sha=$(nix-prefetch-url --type sha256 "$darwin_url")

echo "linux  sha256: $linux_sha"
echo "darwin sha256: $darwin_sha"

# Rewrites land in PR 3 once the prebuilt-fetch attrs exist in
# flake.nix. For now this script just prints what would change.
cat <<EOF
# Apply by editing flake.nix:
version = "$version";
linuxSha = "$linux_sha";
darwinSha = "$darwin_sha";
EOF
