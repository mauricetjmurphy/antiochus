package crypto

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
)

var (
	MagicV1 = []byte("CGRAM1")
	MagicV2 = []byte("CGRAM2")
)

const (
	TypeText byte = 0x01
	TypeFile byte = 0x02
)

// BuildPacket encrypts plaintext and assembles a v2 wire-format packet.
func BuildPacket(payloadType byte, meta map[string]string, plaintext []byte, passphrase string) ([]byte, error) {
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("marshal meta: %w", err)
	}

	salt, aesNonce, chachaNonce, ct, err := Encrypt(plaintext, passphrase)
	if err != nil {
		return nil, err
	}

	metaLen := len(metaJSON)
	packet := make([]byte, 0, 6+1+2+metaLen+SaltLen+AESNonceLen+ChachaNonceLen+len(ct))
	packet = append(packet, MagicV2...)
	packet = append(packet, payloadType)
	packet = binary.BigEndian.AppendUint16(packet, uint16(metaLen))
	packet = append(packet, metaJSON...)
	packet = append(packet, salt...)
	packet = append(packet, aesNonce...)
	packet = append(packet, chachaNonce...)
	packet = append(packet, ct...)

	return packet, nil
}

// ParsePacket parses and decrypts a packet (v1 or v2).
// Returns (payloadType, meta, plaintext).
func ParsePacket(packet []byte, passphrase string) (byte, map[string]string, []byte, error) {
	if len(packet) < 6 {
		return 0, nil, nil, fmt.Errorf("packet too short")
	}

	magic := string(packet[:6])

	switch magic {
	case "CGRAM2":
		return parseV2(packet, passphrase)
	case "CGRAM1":
		return parseV1(packet, passphrase)
	default:
		return 0, nil, nil, fmt.Errorf("invalid packet: bad magic bytes %q", magic)
	}
}

func parseV2(packet []byte, passphrase string) (byte, map[string]string, []byte, error) {
	offset := 6

	if len(packet) < offset+3 {
		return 0, nil, nil, fmt.Errorf("packet too short for v2 header")
	}

	payloadType := packet[offset]
	offset++

	metaLen := int(binary.BigEndian.Uint16(packet[offset : offset+2]))
	offset += 2

	if len(packet) < offset+metaLen+SaltLen+AESNonceLen+ChachaNonceLen {
		return 0, nil, nil, fmt.Errorf("packet too short for v2 payload")
	}

	var meta map[string]string
	if metaLen > 0 {
		if err := json.Unmarshal(packet[offset:offset+metaLen], &meta); err != nil {
			return 0, nil, nil, fmt.Errorf("parse meta: %w", err)
		}
	} else {
		meta = make(map[string]string)
	}
	offset += metaLen

	salt := packet[offset : offset+SaltLen]
	offset += SaltLen
	aesNonce := packet[offset : offset+AESNonceLen]
	offset += AESNonceLen
	chachaNonce := packet[offset : offset+ChachaNonceLen]
	offset += ChachaNonceLen
	ct := packet[offset:]

	plaintext, err := Decrypt(salt, aesNonce, chachaNonce, ct, passphrase)
	if err != nil {
		return 0, nil, nil, err
	}

	return payloadType, meta, plaintext, nil
}

func parseV1(packet []byte, passphrase string) (byte, map[string]string, []byte, error) {
	offset := 6

	if len(packet) < offset+SaltLen+AESNonceLen+ChachaNonceLen+1 {
		return 0, nil, nil, fmt.Errorf("packet too short for v1 payload")
	}

	salt := packet[offset : offset+SaltLen]
	offset += SaltLen
	aesNonce := packet[offset : offset+AESNonceLen]
	offset += AESNonceLen
	chachaNonce := packet[offset : offset+ChachaNonceLen]
	offset += ChachaNonceLen
	ct := packet[offset:]

	plaintext, err := Decrypt(salt, aesNonce, chachaNonce, ct, passphrase)
	if err != nil {
		return 0, nil, nil, err
	}

	return TypeText, make(map[string]string), plaintext, nil
}
