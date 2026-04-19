#!/usr/bin/env bash
# Build Antiochus.app bundle for macOS.
# Requires: dist/antiochus-darwin-{amd64,arm64,universal} + assets/antiochus.icns.
set -euo pipefail

ARCH="${ARCH:-universal}"          # amd64 | arm64 | universal
VERSION="${VERSION:-2.0.0}"
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
APP="$ROOT/dist/Antiochus.app"

case "$ARCH" in
  amd64)     BIN="$ROOT/dist/antiochus-darwin-amd64" ;;
  arm64)     BIN="$ROOT/dist/antiochus-darwin-arm64" ;;
  universal) BIN="$ROOT/dist/antiochus-darwin-universal" ;;
  *) echo "unsupported ARCH=$ARCH"; exit 1 ;;
esac

[[ -f "$BIN" ]] || { echo "missing $BIN — run: make build-darwin-$ARCH"; exit 1; }
[[ -f "$ROOT/assets/antiochus.icns" ]] || { echo "missing assets/antiochus.icns"; exit 1; }

rm -rf "$APP"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"

install -m755 "$BIN" "$APP/Contents/MacOS/antiochus-bin"

# Wrapper so double-click invokes the binary with --open.
cat > "$APP/Contents/MacOS/antiochus" <<'EOF'
#!/usr/bin/env bash
HERE="$(cd "$(dirname "$0")" && pwd)"
exec "$HERE/antiochus-bin" --open
EOF
chmod +x "$APP/Contents/MacOS/antiochus"

install -m644 "$ROOT/assets/antiochus.icns" "$APP/Contents/Resources/antiochus.icns"
sed "s/@VERSION@/$VERSION/g" "$ROOT/packaging/macos/Info.plist.template" > "$APP/Contents/Info.plist"

echo "built: $APP"
echo "note: unsigned bundles trigger Gatekeeper; users must right-click > Open on first launch,"
echo "      or you can codesign & notarize for a clean distribution experience."
