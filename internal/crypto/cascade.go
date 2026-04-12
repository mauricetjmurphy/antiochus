package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	AESNonceLen    = 12
	ChachaNonceLen = 12
)

// Encrypt performs cascade encryption: AES-256-GCM then ChaCha20-Poly1305.
// Returns the raw crypto components (salt, aesNonce, chachaNonce, ciphertext).
func Encrypt(plaintext []byte, passphrase string) (salt, aesNonce, chachaNonce, ciphertext []byte, err error) {
	salt = make([]byte, SaltLen)
	if _, err = rand.Read(salt); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("generate salt: %w", err)
	}

	aesNonce = make([]byte, AESNonceLen)
	if _, err = rand.Read(aesNonce); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("generate AES nonce: %w", err)
	}

	chachaNonce = make([]byte, ChachaNonceLen)
	if _, err = rand.Read(chachaNonce); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("generate ChaCha nonce: %w", err)
	}

	aesKey, chachaKey := DeriveKeys(passphrase, salt)

	// Layer 1: AES-256-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("AES cipher: %w", err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("AES GCM: %w", err)
	}
	aesCT := aesGCM.Seal(nil, aesNonce, plaintext, nil)

	// Layer 2: ChaCha20-Poly1305
	chachaCipher, err := chacha20poly1305.New(chachaKey)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("ChaCha cipher: %w", err)
	}
	ciphertext = chachaCipher.Seal(nil, chachaNonce, aesCT, nil)

	return salt, aesNonce, chachaNonce, ciphertext, nil
}

// Decrypt reverses cascade encryption: ChaCha20-Poly1305 then AES-256-GCM.
func Decrypt(salt, aesNonce, chachaNonce, ciphertext []byte, passphrase string) ([]byte, error) {
	aesKey, chachaKey := DeriveKeys(passphrase, salt)

	// Layer 2: ChaCha20-Poly1305
	chachaCipher, err := chacha20poly1305.New(chachaKey)
	if err != nil {
		return nil, fmt.Errorf("ChaCha cipher: %w", err)
	}
	aesCT, err := chachaCipher.Open(nil, chachaNonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("ChaCha decrypt: %w", err)
	}

	// Layer 1: AES-256-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("AES cipher: %w", err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("AES GCM: %w", err)
	}
	plaintext, err := aesGCM.Open(nil, aesNonce, aesCT, nil)
	if err != nil {
		return nil, fmt.Errorf("AES decrypt: %w", err)
	}

	return plaintext, nil
}
