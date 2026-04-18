package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Session holds the server-side ECDH state, shared key, and decrypted passphrase.
type Session struct {
	mu         sync.RWMutex
	privateKey *ecdh.PrivateKey
	publicKey  *ecdh.PublicKey
	passphrase string
	sharedKey  []byte // 32-byte AES key derived from ECDH — used for WS encryption
}

func NewSession() *Session {
	return &Session{}
}

// GenerateKeyPair creates a new ECDH P-256 keypair for this session.
func (s *Session) GenerateKeyPair() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	curve := ecdh.P256()
	priv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate ECDH key: %w", err)
	}
	s.privateKey = priv
	s.publicKey = priv.PublicKey()
	return nil
}

// PublicKeyBase64 returns the server's public key as base64.
func (s *Session) PublicKeyBase64() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return base64.StdEncoding.EncodeToString(s.publicKey.Bytes())
}

// DecryptPassphrase takes the client's public key and encrypted passphrase,
// derives the shared secret via ECDH, decrypts the passphrase, and stores
// the shared key for future WS encryption.
func (s *Session) DecryptPassphrase(clientPubKeyB64, encryptedB64, ivB64 string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.privateKey == nil {
		return fmt.Errorf("no keypair generated — call /api/session/key first")
	}

	clientPubBytes, err := base64.StdEncoding.DecodeString(clientPubKeyB64)
	if err != nil {
		return fmt.Errorf("decode client public key: %w", err)
	}

	curve := ecdh.P256()
	clientPub, err := curve.NewPublicKey(clientPubBytes)
	if err != nil {
		return fmt.Errorf("parse client public key: %w", err)
	}

	shared, err := s.privateKey.ECDH(clientPub)
	if err != nil {
		return fmt.Errorf("ECDH: %w", err)
	}

	// P-256 ECDH produces 32 bytes — use as AES-256 key
	aesKey := shared

	iv, err := base64.StdEncoding.DecodeString(ivB64)
	if err != nil {
		return fmt.Errorf("decode IV: %w", err)
	}
	encrypted, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return fmt.Errorf("decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return fmt.Errorf("AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("AES GCM: %w", err)
	}
	plaintext, err := gcm.Open(nil, iv, encrypted, nil)
	if err != nil {
		return fmt.Errorf("decrypt passphrase: %w", err)
	}

	s.passphrase = string(plaintext)
	s.sharedKey = aesKey // Keep for WS encryption

	// Discard the ECDH keypair — single use
	s.privateKey = nil
	s.publicKey = nil

	return nil
}

// EncryptJSON encrypts a JSON-serializable value with AES-256-GCM using the
// session shared key. Returns base64(iv) + "." + base64(ciphertext).
func (s *Session) EncryptJSON(data interface{}) (string, error) {
	s.mu.RLock()
	key := s.sharedKey
	s.mu.RUnlock()

	if key == nil {
		return "", fmt.Errorf("no session key — handshake not completed")
	}

	plaintext, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("AES GCM: %w", err)
	}

	iv := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("generate IV: %w", err)
	}

	ciphertext := gcm.Seal(nil, iv, plaintext, nil)

	return base64.StdEncoding.EncodeToString(iv) + "." + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// HasSharedKey returns whether the WS encryption key is available.
func (s *Session) HasSharedKey() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sharedKey != nil
}

// Passphrase returns the stored passphrase.
func (s *Session) Passphrase() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.passphrase
}

// HasPassphrase returns whether a passphrase has been set.
func (s *Session) HasPassphrase() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.passphrase != ""
}

// Clear wipes all session state from memory.
func (s *Session) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.passphrase = ""
	s.sharedKey = nil
	s.privateKey = nil
	s.publicKey = nil
}

// ── HTTP Handlers ──────────────────────────────────────────────

func (h *Handler) GetSessionKey(w http.ResponseWriter, r *http.Request) {
	if err := h.Session.GenerateKeyPair(); err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"public_key": h.Session.PublicKeyBase64(),
	})
}

func (h *Handler) StartPoll(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClientPublicKey     string `json:"client_public_key"`
		EncryptedPassphrase string `json:"encrypted_passphrase"`
		IV                  string `json:"iv"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if req.ClientPublicKey == "" || req.EncryptedPassphrase == "" || req.IV == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "client_public_key, encrypted_passphrase, and iv required"})
		return
	}

	if err := h.Session.DecryptPassphrase(req.ClientPublicKey, req.EncryptedPassphrase, req.IV); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "key exchange failed: " + err.Error()})
		return
	}

	if h.Cfg.Telegram.BotToken == "" {
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true", "polling": "false"})
		return
	}

	if h.OnStartPoll != nil {
		if err := h.OnStartPoll(h.Session.Passphrase()); err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	WriteJSON(w, http.StatusOK, map[string]string{"ok": "true", "polling": "true"})
}

func (h *Handler) StopPoll(w http.ResponseWriter, r *http.Request) {
	if h.OnStopPoll != nil {
		h.OnStopPoll()
	}
	h.Session.Clear()
	WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

// FetchUpdates triggers a one-shot getUpdates call on the poller.
func (h *Handler) FetchUpdates(w http.ResponseWriter, r *http.Request) {
	if h.OnFetchUpdates == nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "fetch not wired up"})
		return
	}
	count, err := h.OnFetchUpdates()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	WriteJSON(w, http.StatusOK, map[string]int{"processed": count})
}
