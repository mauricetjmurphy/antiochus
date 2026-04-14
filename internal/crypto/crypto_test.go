package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKeys(t *testing.T) {
	salt := make([]byte, SaltLen)
	aesKey, chachaKey := DeriveKeys("test-passphrase", salt)

	if len(aesKey) != 32 {
		t.Fatalf("expected 32-byte AES key, got %d", len(aesKey))
	}
	if len(chachaKey) != 32 {
		t.Fatalf("expected 32-byte ChaCha key, got %d", len(chachaKey))
	}
	if bytes.Equal(aesKey, chachaKey) {
		t.Fatal("AES and ChaCha keys should differ")
	}
}

func TestCascadeRoundTrip(t *testing.T) {
	plaintext := []byte("Antiochus self-test payload")
	passphrase := "test-passphrase-42"

	salt, aesNonce, chachaNonce, ct, err := Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	decrypted, err := Decrypt(salt, aesNonce, chachaNonce, ct, passphrase)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("round-trip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestWrongPassphrase(t *testing.T) {
	plaintext := []byte("secret message")
	salt, aesNonce, chachaNonce, ct, err := Encrypt(plaintext, "correct")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	_, err = Decrypt(salt, aesNonce, chachaNonce, ct, "wrong")
	if err == nil {
		t.Fatal("expected error with wrong passphrase")
	}
}

func TestPacketV2TextRoundTrip(t *testing.T) {
	plaintext := []byte("hello from v2")
	passphrase := "test-pass"

	packet, err := BuildPacket(TypeText, map[string]string{}, plaintext, passphrase)
	if err != nil {
		t.Fatalf("build packet: %v", err)
	}

	if string(packet[:6]) != "ANTIO2" {
		t.Fatalf("expected ANTIO2 magic, got %q", packet[:6])
	}

	ptype, meta, decrypted, err := ParsePacket(packet, passphrase)
	if err != nil {
		t.Fatalf("parse packet: %v", err)
	}

	if ptype != TypeText {
		t.Fatalf("expected TypeText, got %d", ptype)
	}
	if len(meta) != 0 {
		t.Fatalf("expected empty meta, got %v", meta)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("round-trip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestPacketV2FileRoundTrip(t *testing.T) {
	fileData := []byte("\x89PNG\r\n\x1a\nfake image data")
	passphrase := "file-pass"
	meta := map[string]string{"filename": "test.png"}

	packet, err := BuildPacket(TypeFile, meta, fileData, passphrase)
	if err != nil {
		t.Fatalf("build packet: %v", err)
	}

	ptype, gotMeta, decrypted, err := ParsePacket(packet, passphrase)
	if err != nil {
		t.Fatalf("parse packet: %v", err)
	}

	if ptype != TypeFile {
		t.Fatalf("expected TypeFile, got %d", ptype)
	}
	if gotMeta["filename"] != "test.png" {
		t.Fatalf("expected filename test.png, got %s", gotMeta["filename"])
	}
	if !bytes.Equal(fileData, decrypted) {
		t.Fatalf("round-trip failed")
	}
}

func TestPacketV1BackwardCompat(t *testing.T) {
	// Manually build a V1 packet
	plaintext := []byte("legacy message")
	passphrase := "v1-pass"

	salt, aesNonce, chachaNonce, ct, err := Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	// Assemble V1 format: MAGIC(6) + salt(32) + aesNonce(12) + chachaNonce(12) + ct
	v1Packet := make([]byte, 0)
	v1Packet = append(v1Packet, MagicV1...)
	v1Packet = append(v1Packet, salt...)
	v1Packet = append(v1Packet, aesNonce...)
	v1Packet = append(v1Packet, chachaNonce...)
	v1Packet = append(v1Packet, ct...)

	ptype, _, decrypted, err := ParsePacket(v1Packet, passphrase)
	if err != nil {
		t.Fatalf("parse v1 packet: %v", err)
	}

	if ptype != TypeText {
		t.Fatalf("v1 should always be TypeText, got %d", ptype)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("v1 round-trip failed")
	}
}
