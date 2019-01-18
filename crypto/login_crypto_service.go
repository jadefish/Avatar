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

// LoginCryptoService is a cryptography service compatible with modern
// "classic" (2D) clients.
type LoginCryptoService struct {
	seed           avatar.Seed
	maskLo, maskHi uint32
	keyLo, keyHi   uint32
}

// NewLoginCryptoService creates a new cryptography service.
func NewLoginCryptoService(
	seed avatar.Seed,
	version avatar.ClientVersion,
) (*LoginCryptoService, error) {
	maskLo := computeMaskLo(seed)
	maskHi := computeMaskHi(seed)
	keyPair, err := crypto.GetClientKeyPair(version)

	if err != nil {
		return nil, errors.Wrap(err, "cryptography service creation")
	}

	return &LoginCryptoService{
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
func (c LoginCryptoService) GetSeed() avatar.Seed {
	return c.seed
}

// GetMasks returns the pair of client masks.
func (c LoginCryptoService) GetMasks() avatar.KeyPair {
	return avatar.KeyPair{
		Lo: c.maskLo,
		Hi: c.maskHi,
	}
}

// GetKeys returns the pair of version-specific client keys.
func (c LoginCryptoService) GetKeys() avatar.KeyPair {
	return avatar.KeyPair{
		Lo: c.keyLo,
		Hi: c.keyHi,
	}
}

// Encrypt encrypts src, returning the encrypted data.
// The length of the returned slice will always equal the length of the
// source slice.
func (c LoginCryptoService) Encrypt(src []byte) ([]byte, error) {
	return loginCrypt(c, src)
}

// Decrypt src, returning the unencrypted data
func (c LoginCryptoService) Decrypt(src []byte) ([]byte, error) {
	return loginCrypt(c, src)
}

// loginCrypt is capable of both encrypting and decrypting login packets.
func loginCrypt(cs LoginCryptoService, src []byte) ([]byte, error) {
	dest := make([]byte, len(src))
	lo := cs.maskLo
	hi := cs.maskHi

	for i := range src {
		dest[i] = src[i] ^ byte(lo)

		maskLo := lo
		maskHi := hi

		lo = ((maskLo >> 1) | (maskHi << 31)) ^ cs.keyLo
		maskHi = ((maskHi >> 1) | (maskLo << 31)) ^ cs.keyHi
		hi = ((maskHi >> 1) | (maskLo << 31)) ^ cs.keyHi
	}

	return dest, nil
}
