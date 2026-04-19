VERSION := 2.0.0
FRONTEND_DIR := frontend
BINARY := antiochus
LDFLAGS     := -s -w -X main.version=$(VERSION)
LDFLAGS_WIN := -s -w -H=windowsgui -X main.version=$(VERSION)

.PHONY: all frontend backend clean dev-backend dev-frontend release build-all test \
        install-linux uninstall-linux appimage appimage-build appimage-arm64 \
        uninstall-appimage build-darwin-universal macos-app windows-resources

APPIMAGE_HOME      := $(HOME)/Applications/Antiochus
APPIMAGE_MENU_DIR  := $(HOME)/.local/share/applications
APPIMAGE_INSTALLED := $(APPIMAGE_HOME)/Antiochus.AppImage
APPIMAGE_ICON      := $(APPIMAGE_HOME)/$(BINARY).png

PREFIX ?= /usr/local
DESTDIR ?=

all: frontend backend

# ── Frontend ──────────────────────────────────
frontend:
	cd $(FRONTEND_DIR) && npm ci && npm run build

# ── Backend (embeds frontend/dist) ────────────
backend:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY) .

# ── Tests ─────────────────────────────────────
test:
	go test ./internal/crypto/... -v -count=1

# ── Cross-compile ─────────────────────────────
build-linux-amd64: frontend
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64 .

build-linux-arm64: frontend
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-arm64 .

build-darwin-amd64: frontend
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-amd64 .

build-darwin-arm64: frontend
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-arm64 .

build-windows-386: frontend
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags "$(LDFLAGS_WIN)" -o dist/$(BINARY)-windows-x86.exe .

# amd64 picks up resource_windows_amd64.syso (icon + version info) from repo root.
build-windows-amd64: frontend windows-resources
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS_WIN)" -o dist/$(BINARY)-windows-amd64.exe .

build-windows-arm64: frontend
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS_WIN)" -o dist/$(BINARY)-windows-arm64.exe .

# Universal darwin binary (Intel + Apple Silicon) via lipo.
build-darwin-universal: build-darwin-amd64 build-darwin-arm64
	lipo -create -output dist/$(BINARY)-darwin-universal \
		dist/$(BINARY)-darwin-amd64 \
		dist/$(BINARY)-darwin-arm64

release: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 build-windows-arm64

build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-386 build-windows-amd64 build-windows-arm64

# ── Development ───────────────────────────────
dev-backend:
	go run . --addr 127.0.0.1:8080

dev-frontend:
	cd $(FRONTEND_DIR) && npm run dev

# ── Icon normalization ────────────────────────
# Accepts any of png/svg/webp/jpg/jpeg in assets/ and produces a 256x256 PNG.
# Requires ImageMagick (magick or convert).
dist/$(BINARY)-icon-256.png:
	@mkdir -p dist
	@src=""; for ext in png svg webp jpg jpeg; do \
		if [ -f "assets/$(BINARY).$$ext" ]; then src="assets/$(BINARY).$$ext"; break; fi; \
	done; \
	if [ -z "$$src" ]; then echo "no assets/$(BINARY).{png,svg,webp,jpg,jpeg} found"; exit 1; fi; \
	if command -v magick >/dev/null 2>&1; then IM=magick; \
	elif command -v convert >/dev/null 2>&1; then IM=convert; \
	else echo "ImageMagick required (install imagemagick)"; exit 1; fi; \
	echo "icon: $$src → $@ (256x256)"; \
	$$IM "$$src" -resize 256x256^ -gravity center -extent 256x256 "$@"

# ── Linux desktop install ─────────────────────
install-linux: build-linux-amd64 dist/$(BINARY)-icon-256.png
	install -Dm755 dist/$(BINARY)-linux-amd64 $(DESTDIR)$(PREFIX)/bin/$(BINARY)
	install -Dm644 packaging/linux/$(BINARY).desktop $(DESTDIR)$(PREFIX)/share/applications/$(BINARY).desktop
	install -Dm644 dist/$(BINARY)-icon-256.png $(DESTDIR)$(PREFIX)/share/icons/hicolor/256x256/apps/$(BINARY).png
	-update-desktop-database $(DESTDIR)$(PREFIX)/share/applications 2>/dev/null || true
	-gtk-update-icon-cache -f $(DESTDIR)$(PREFIX)/share/icons/hicolor 2>/dev/null || true

uninstall-linux:
	rm -f $(DESTDIR)$(PREFIX)/bin/$(BINARY)
	rm -f $(DESTDIR)$(PREFIX)/share/applications/$(BINARY).desktop
	rm -f $(DESTDIR)$(PREFIX)/share/icons/hicolor/256x256/apps/$(BINARY).png
	-update-desktop-database $(DESTDIR)$(PREFIX)/share/applications 2>/dev/null || true
	-gtk-update-icon-cache -f $(DESTDIR)$(PREFIX)/share/icons/hicolor 2>/dev/null || true

# ── AppImage ──────────────────────────────────
# Build-only — produces dist/Antiochus-$(VERSION)-x86_64.AppImage for distribution.
appimage-build: build-linux-amd64
	ARCH=x86_64 VERSION=$(VERSION) packaging/linux/build-appimage.sh

# Default: build AND install into $(APPIMAGE_HOME). The .AppImage goes straight
# to ~/Applications/Antiochus/ — nothing final lands in dist/.
appimage: build-linux-amd64 dist/$(BINARY)-icon-256.png
	@mkdir -p $(APPIMAGE_HOME) $(APPIMAGE_MENU_DIR)
	ARCH=x86_64 VERSION=$(VERSION) OUT=$(APPIMAGE_INSTALLED) \
		packaging/linux/build-appimage.sh
	install -m644 dist/$(BINARY)-icon-256.png $(APPIMAGE_ICON)
	sed -e 's|^Exec=.*|Exec=$(APPIMAGE_INSTALLED) --open|' \
	    -e 's|^Icon=.*|Icon=$(APPIMAGE_ICON)|' \
	    packaging/linux/$(BINARY).desktop > $(APPIMAGE_MENU_DIR)/$(BINARY).desktop
	-update-desktop-database $(APPIMAGE_MENU_DIR) 2>/dev/null || true
	@echo
	@echo "Installed: $(APPIMAGE_HOME)/"
	@echo "Launcher:  $(APPIMAGE_MENU_DIR)/$(BINARY).desktop"
	@echo "Search 'Antiochus' in your app menu, then right-click → Pin to Dash."

# Cross-arch build (not installed — can't run arm64 on x86_64 and vice versa).
appimage-arm64: build-linux-arm64
	ARCH=aarch64 VERSION=$(VERSION) packaging/linux/build-appimage.sh

uninstall-appimage:
	rm -rf $(APPIMAGE_HOME)
	rm -f  $(APPIMAGE_MENU_DIR)/$(BINARY).desktop
	-update-desktop-database $(APPIMAGE_MENU_DIR) 2>/dev/null || true
	@echo "Removed $(APPIMAGE_HOME) and launcher."

# ── macOS .app bundle ─────────────────────────
# Default: universal (Intel + Apple Silicon). Override with ARCH=amd64 or arm64.
macos-app: build-darwin-universal
	ARCH=universal VERSION=$(VERSION) packaging/macos/build-app.sh

# ── Windows resources (icon + version info) ───
# Generates resource_windows_amd64.syso which Go auto-includes in windows/amd64 builds.
windows-resources:
	@command -v goversioninfo >/dev/null 2>&1 || \
		go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
	@[ -f assets/$(BINARY).ico ] || { echo "missing assets/$(BINARY).ico"; exit 1; }
	goversioninfo -64 -o resource_windows_amd64.syso packaging/windows/versioninfo.json

# ── Clean ─────────────────────────────────────
clean:
	rm -rf dist/ $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/node_modules
	rm -rf dist/AppDir dist/Antiochus.app
	rm -f resource_windows_amd64.syso
