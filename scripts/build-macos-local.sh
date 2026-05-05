#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

arch="${1:-$(go env GOARCH)}"
case "$arch" in
  arm64|amd64) ;;
  *)
    echo "unsupported macOS architecture: $arch (expected arm64 or amd64)" >&2
    exit 2
    ;;
esac

if [[ "$(uname -s)" != "Darwin" ]]; then
  cat >&2 <<EOF
This app uses Fyne/CGO and must be built for macOS with Apple's SDK.

Run this script on a Mac:
  scripts/build-macos-local.sh $arch

From Linux, plain GOOS=darwin is not enough. A Docker cross-build also needs
a macOS SDK and a Darwin-capable linker, for example:
  go run github.com/fyne-io/fyne-cross@latest darwin \\
    -arch=$arch \\
    -macosx-sdk-path=/path/to/MacOSX.sdk \\
    -env GOTOOLCHAIN=auto \\
    -tags netgo
EOF
  exit 1
fi

version="$(grep -oE 'Version = "[^"]+"' FyneApp.toml | sed -E 's/Version = "([^"]+)"/\1/')"
build_dir="$root/dist/macos-$arch"
raw_bin="$root/dist/cosmonaut-darwin-$arch"
app_dir="$build_dir/Cosmonaut.app"
zip_path="$root/dist/cosmonaut-macos-$arch.zip"

rm -rf "$build_dir" "$raw_bin" "$zip_path"
mkdir -p "$app_dir/Contents/MacOS" "$app_dir/Contents/Resources"

echo "building cosmonaut $version for darwin/$arch"
GOOS=darwin GOARCH="$arch" CGO_ENABLED=1 go build \
  -tags netgo \
  -trimpath \
  -ldflags "-s -w" \
  -o "$raw_bin" \
  .

cp "$raw_bin" "$app_dir/Contents/MacOS/cosmonaut"
cp "$root/dist/Info.plist" "$app_dir/Contents/Info.plist"
cp "$root/assets/logo.icns" "$app_dir/Contents/Resources/icon.icns"
cp "$root/dist/cosmonaut.config.example.json" "$build_dir/cosmonaut.config.example.json"

if command -v codesign >/dev/null 2>&1; then
  codesign --force --deep --sign - "$app_dir" >/dev/null
fi

if command -v ditto >/dev/null 2>&1; then
  (cd "$build_dir" && ditto -c -k --sequesterRsrc --keepParent Cosmonaut.app "$zip_path")
else
  (cd "$build_dir" && zip -qr "$zip_path" Cosmonaut.app)
fi

echo "raw binary: $raw_bin"
echo "app bundle zip: $zip_path"
