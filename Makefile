VERSION := 2.0.0
FRONTEND_DIR := frontend
BINARY := antiochus
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all frontend backend clean dev-backend dev-frontend release build-all test

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
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-windows-x86.exe .

build-windows-amd64: frontend
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-windows-amd64.exe .

build-windows-arm64: frontend
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-windows-arm64.exe .

release: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 build-windows-arm64

build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-386 build-windows-amd64 build-windows-arm64

# ── Development ───────────────────────────────
dev-backend:
	go run . --addr 127.0.0.1:8080

dev-frontend:
	cd $(FRONTEND_DIR) && npm run dev

# ── Clean ─────────────────────────────────────
clean:
	rm -rf dist/ $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/node_modules
