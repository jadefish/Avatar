package crypto

import (
	"github.com/jadefish/avatar"
	"github.com/jadefish/avatar/internal/crypto"
	"github.com/pkg/errors"
)

const (
	loMask1 = 0x00001357
	loMask2 = 0xffffaaaa
	loMask3 = 0x0000ffff
	hiMask1 = 0x43210000
	hiMask2 = 0xabcdffff
	hiMask3 = 0xffff0000
)

// ClassicCryptoService is a cryptography service compatible with modern
// "classic" (2D) clients.
type ClassicCryptoService struct {
	seed           avatar.Seed
	maskLo, maskHi uint32
	keyLo, keyHi   uint32
}

// NewClassicCryptoService creates a new cryptography service.
func NewClassicCryptoService(
	seed avatar.Seed,
	version avatar.ClientVersion,
) (*ClassicCryptoService, error) {
	maskLo := computeMaskLo(seed)
	maskHi := computeMaskHi(seed)
	keyPair, err := crypto.GetClientKeyPair(version)

	if err != nil {
		return nil, errors.Wrap(err, "cryptography service creation")
	}

	return &ClassicCryptoService{
		seed:   seed,
		maskLo: maskLo,
		maskHi: maskHi,
		keyLo:  keyPair.Lo,
		keyHi:  keyPair.Hi,
	}, nil
}

func computeMaskLo(seed avatar.Seed) uint32 {
	s := uint32(seed)

	return ((^s ^ loMask1) << 16) | ((s ^ loMask2) & loMask3)
}

func computeMaskHi(seed avatar.Seed) uint32 {
	s := uint32(seed)

	return ((s ^ hiMask1) >> 16) | ((^s ^ hiMask2) & hiMask3)
}

// GetSeed returns the client's seed.
func (c ClassicCryptoService) GetSeed() avatar.Seed {
	return c.seed
}

// GetMasks returns the pair of client masks.
func (c ClassicCryptoService) GetMasks() avatar.KeyPair {
	return avatar.KeyPair{
		Lo: c.maskLo,
		Hi: c.maskHi,
	}
}

// GetKeys returns the pair of version-specific client keys.
func (c ClassicCryptoService) GetKeys() avatar.KeyPair {
	return avatar.KeyPair{
		Lo: c.keyLo,
		Hi: c.keyHi,
	}
}

// Encrypt data using the client's cryptography facilities.
func (c ClassicCryptoService) Encrypt(data []byte) ([]byte, error) {
	// TODO
	return []byte{0x00}, nil
}

// Decrypt data using the client's cryptography facilities.
func (c ClassicCryptoService) Decrypt(data []byte) ([]byte, error) {
	// TODO
	return []byte{0x00}, nil
}

// LoginDecrypt decrypts the client's login data.
func (c ClassicCryptoService) LoginDecrypt(src []byte) ([]byte, error) {
	dest := make([]byte, len(src))

	for i := range src {
		dest[i] = src[i] ^ byte(c.maskLo)

		maskLo := c.maskLo
		maskHi := c.maskHi

		c.maskLo = ((maskLo >> 1) | (maskHi << 31)) ^ c.keyLo
		maskHi = ((maskHi >> 1) | (maskLo << 31)) ^ c.keyHi
		c.maskHi = ((maskHi >> 1) | (maskLo << 31)) ^ c.keyHi
	}

	return dest, nil
}
