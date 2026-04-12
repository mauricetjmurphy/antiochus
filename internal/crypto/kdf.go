package crypto

import (
	"golang.org/x/crypto/argon2"
)

const (
	SaltLen      = 32
	Argon2KeyLen = 64 // 32 bytes per cipher
)

// Default Argon2id parameters (used when no config is provided).
var DefaultKDFParams = KDFParams{
	TimeCost:    4,
	MemoryKB:    262144, // 256 MB
	Parallelism: 4,
}

// KDFParams holds tunable Argon2id parameters.
type KDFParams struct {
	TimeCost    uint32
	MemoryKB    uint32
	Parallelism uint8
}

// ActiveParams is the global KDF configuration. Set from config at startup.
var ActiveParams = DefaultKDFParams

// DeriveKeys derives two 256-bit keys from a passphrase using Argon2id.
// Returns (aesKey, chachaKey).
func DeriveKeys(passphrase string, salt []byte) (aesKey, chachaKey []byte) {
	raw := argon2.IDKey(
		[]byte(passphrase),
		salt,
		ActiveParams.TimeCost,
		ActiveParams.MemoryKB,
		ActiveParams.Parallelism,
		Argon2KeyLen,
	)
	aesKey = raw[:32]
	chachaKey = raw[32:64]
	return
}
