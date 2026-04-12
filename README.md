# Antiochus

Encrypted messaging over Telegram with a WhatsApp-like web UI, compiled into a single binary.

Both parties run the same binary on their machines. Messages are cascade-encrypted (AES-256-GCM + ChaCha20-Poly1305) before they touch Telegram's servers — the transport is untrusted, only the endpoints matter.

**New here?** Read the [Setup Guide](docs/setup-guide.md) to get started.

## How It Works

```
You                          Telegram                        Friend
 |                              |                              |
 |  plaintext                   |                              |
 |  -> Argon2id KDF             |                              |
 |  -> AES-256-GCM              |                              |
 |  -> ChaCha20-Poly1305        |                              |
 |  -> base64/document    ---->  |  (encrypted blob)     ---->  |
 |                              |                              |  -> ChaCha20 decrypt
 |                              |                              |  -> AES-GCM decrypt
 |                              |                              |  -> plaintext
```

- 512-bit effective key material (two independent 256-bit keys)
- Argon2id KDF: 256 MB memory, 4 iterations, 4 threads
- Random salt + nonces per message
- Passphrase never stored on disk

## Quick Start

### Prerequisites

- Go 1.23+
- Node.js 18+
- A Telegram bot token from [@BotFather](https://t.me/BotFather)

### Build

```bash
make all
```

This builds the frontend (Vite + React + Tailwind) and compiles the Go binary with the UI embedded:

```
dist/antiochus    # ~7.5 MB single binary
```

### Run

```bash
./dist/antiochus
```

Open http://127.0.0.1:8080 in your browser. Enter a passphrase, configure your bot token in Settings, add friends, and start chatting.

A setup guide is available at http://127.0.0.1:8080/guide (also linked from the login screen).

### Give It to a Friend

Build for their platform and send them the binary:

```bash
make release
```

Produces:
```
dist/antiochus-linux-amd64
dist/antiochus-linux-arm64
dist/antiochus-darwin-amd64
dist/antiochus-darwin-arm64
dist/antiochus-windows-amd64.exe
dist/antiochus-windows-arm64.exe
```

No dependencies to install — it's a single file. They run it, open the browser, set the same bot token and passphrase, and you're connected.

## Configuration

Config file: `~/.ciphergram/ciphergram.yml` (also checks `internal/config/prod.yml` in the working directory first)

```yaml
server:
  addr: "127.0.0.1:8080"

telegram:
  bot_token: ""              # from @BotFather
  chat_id: ""                # optional default recipient
  poll_timeout: 30
  max_file_size_mb: 45
  message_char_limit: 4000

crypto:
  argon2_time_cost: 4        # changing these breaks compatibility
  argon2_memory_mb: 256      # with anyone using different values
  argon2_parallelism: 4

storage:
  max_messages_per_friend: 500

friends:
  alice: "123456789"
  bob: "987654321"
```

The bot token and friends can also be configured from the web UI (Settings gear + Add Friend button).

### CLI Flags

```
--addr    Override listen address (e.g. --addr 0.0.0.0:9090)
--version Show version
```

## Features

- **Cascade encryption** — AES-256-GCM + ChaCha20-Poly1305 (both must be broken to read a message)
- **ECDH key exchange** — passphrase encrypted via P-256 ECDH before transmission, never sent in cleartext
- **Encrypted WebSocket** — real-time messages encrypted with AES-256-GCM using the session shared key
- **Session authentication** — all API endpoints require an established session (ECDH handshake)
- **File transfer** — send encrypted files up to 45 MB via Telegram documents
- **Friends list** — manage contacts by name, add/remove from the UI or config
- **Who Am I** — auto-discovers your Telegram chat ID in Settings
- **Backward compatible** — decrypts both CGRAM1 (v1) and CGRAM2 (v2) wire formats
- **Cross-platform** — single binary for Linux, macOS, and Windows (amd64 + arm64)
- **Setup guide** — built-in guide served at `/guide`, linked from the login screen

## Security

### What's protected

| Layer | Protection |
|-------|-----------|
| Messages in transit (Telegram) | Cascade encryption: AES-256-GCM + ChaCha20-Poly1305 |
| Passphrase (browser → server) | ECDH P-256 key exchange + AES-256-GCM, sent once |
| WebSocket messages | AES-256-GCM encrypted with ECDH session key |
| API endpoints | Session authentication middleware (401 without handshake) |
| WebSocket connections | Session check before upgrade, origin restricted to localhost |
| Received files | Filename sanitized (path traversal blocked), written with 0600 permissions |
| Config file | Written with 0600 permissions |
| Passphrase storage | In server memory only while active, never written to disk |
| Message history | In-memory only, lost on restart |

### What Telegram sees

- Encrypted text blobs or binary `.enc` files — cannot read message content
- File captions show the filename (e.g. `📎 report.pdf`) but file contents are encrypted
- Message metadata (who sent to whom, when) is visible to Telegram

### Threat model

- The binary binds to `127.0.0.1` by default — binding to `0.0.0.0` prints a warning
- WebSocket origin is restricted to `localhost` and `127.0.0.1` (blocks cross-site hijacking)
- Both parties must share the same passphrase out of band (in person, phone call, etc.)
- No forward secrecy — the same passphrase is reused across messages

## Development

Run the backend and frontend dev servers separately for hot reload:

```bash
# Terminal 1
make dev-backend

# Terminal 2
make dev-frontend
```

The Vite dev server at http://localhost:5173 proxies API calls to the Go backend at :8080.

### Tests

```bash
make test
```

Runs crypto round-trip tests including cascade encrypt/decrypt, wrong passphrase rejection, file packet handling, and CGRAM1 backward compatibility.

## Project Structure

```
antiochus/
├── main.go                     # Entrypoint, go:embed, CLI flags
├── Makefile                    # Build targets
├── docs/
│   ├── setup-guide.md          # User-facing setup guide (embedded in binary)
│   └── implementation-plan.md  # Architecture decisions
├── internal/
│   ├── api/
│   │   ├── handlers/           # HTTP handlers (config, messages, crypto, whoami, session)
│   │   ├── middleware/         # Session authentication middleware
│   │   └── server/             # Chi router, WebSocket hub, polling, guide renderer
│   ├── config/                 # YAML config load/save
│   ├── crypto/                 # Argon2id KDF, AES+ChaCha cascade, wire format
│   ├── models/                 # Shared types (Message, WSEvent, etc.)
│   └── telegram/               # Bot API client, long-poll goroutine, file handling
└── frontend/                   # Vite + React + TypeScript + Tailwind
    └── src/
        ├── api.ts              # ECDH key exchange, REST client, WS decryption
        ├── components/         # PassphraseGate, Sidebar, ChatPanel, Settings, etc.
        └── hooks/              # useWebSocket, useMessages
```

## Wire Format

### CGRAM2 (current)
```
MAGIC "CGRAM2" (6B) | type (1B) | meta_len (2B) | meta (JSON)
| salt (32B) | aes_nonce (12B) | chacha_nonce (12B) | ciphertext
```

Type `0x01` = text, `0x02` = file. Meta contains `{"filename":"..."}` for files.

### CGRAM1 (legacy, read-only)
```
MAGIC "CGRAM1" (6B) | salt (32B) | aes_nonce (12B) | chacha_nonce (12B) | ciphertext
```
