#!/usr/bin/env bash
# Build an AppImage for antiochus.
# Requires: the linux binary already built in dist/, a source icon at
# assets/antiochus.{png,jpg,jpeg,webp,svg}, and the .desktop file.
# The icon is converted + resized to 256x256 PNG via ImageMagick (AppImage
# requires PNG/SVG/XPM — JPG is not accepted).
set -euo pipefail

ARCH="${ARCH:-x86_64}"              # x86_64 | aarch64
VERSION="${VERSION:-2.0.0}"
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
APPDIR="$ROOT/dist/AppDir"
TOOL="$ROOT/dist/appimagetool-$ARCH.AppImage"

case "$ARCH" in
  x86_64)  BIN="$ROOT/dist/antiochus-linux-amd64" ;;
  aarch64) BIN="$ROOT/dist/antiochus-linux-arm64" ;;
  *) echo "unsupported ARCH=$ARCH"; exit 1 ;;
esac

# Resolve source icon in order of preference.
ICON_SRC=""
for ext in png svg webp jpg jpeg; do
  if [[ -f "$ROOT/assets/antiochus.$ext" ]]; then
    ICON_SRC="$ROOT/assets/antiochus.$ext"
    break
  fi
done

[[ -f "$BIN" ]] || { echo "missing $BIN — run: make build-linux-amd64 (or arm64)"; exit 1; }
[[ -n "$ICON_SRC" ]] || { echo "missing assets/antiochus.{png,svg,webp,jpg,jpeg}"; exit 1; }
[[ -f "$ROOT/packaging/linux/antiochus.desktop" ]] || { echo "missing packaging/linux/antiochus.desktop"; exit 1; }

if   command -v magick  >/dev/null 2>&1; then IM=magick
elif command -v convert >/dev/null 2>&1; then IM=convert
else echo "ImageMagick required (install imagemagick) to resize $ICON_SRC → 256x256 PNG"; exit 1
fi

rm -rf "$APPDIR"
mkdir -p "$APPDIR/usr/bin" \
         "$APPDIR/usr/share/applications" \
         "$APPDIR/usr/share/icons/hicolor/256x256/apps"

ICON_DST="$APPDIR/usr/share/icons/hicolor/256x256/apps/antiochus.png"

install -m755 "$BIN"                           "$APPDIR/usr/bin/antiochus"
install -m644 "$ROOT/packaging/linux/antiochus.desktop" "$APPDIR/usr/share/applications/antiochus.desktop"

echo "icon: converting $ICON_SRC → 256x256 PNG"
"$IM" "$ICON_SRC" -resize 256x256^ -gravity center -extent 256x256 "$ICON_DST"
chmod 644 "$ICON_DST"

# AppImage expects the .desktop + icon at AppDir root as well.
cp "$APPDIR/usr/share/applications/antiochus.desktop" "$APPDIR/antiochus.desktop"
cp "$ICON_DST" "$APPDIR/antiochus.png"

cat > "$APPDIR/AppRun" <<'EOF'
#!/usr/bin/env bash
HERE="$(dirname "$(readlink -f "$0")")"
exec "$HERE/usr/bin/antiochus" "$@"
EOF
chmod +x "$APPDIR/AppRun"

if [[ ! -x "$TOOL" ]]; then
  echo "downloading appimagetool..."
  curl -L -o "$TOOL" \
    "https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-$ARCH.AppImage"
  chmod +x "$TOOL"
fi

OUT="${OUT:-$ROOT/dist/Antiochus-$VERSION-$ARCH.AppImage}"
mkdir -p "$(dirname "$OUT")"
ARCH="$ARCH" "$TOOL" "$APPDIR" "$OUT"

# Staging scaffolding isn't useful once the AppImage exists.
rm -rf "$APPDIR"

echo "built: $OUT"
