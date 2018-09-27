package login

import (
	"github.com/jadefish/avatar"
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

type CryptoService struct {
	seed           uint32
	maskLo, maskHi uint32
	keyLo, keyHi   uint32
}

func NewCrypto(seed uint32, version *avatar.ClientVersion) (*CryptoService, error) {
	maskLo := computeMaskLo(seed)
	maskHi := computeMaskHi(seed)
	keyPair, err := getClientKeyPair(version)

	if err != nil {
		return nil, err
	}

	return &CryptoService{
		seed:   seed,
		maskLo: maskLo,
		maskHi: maskHi,
		keyLo:  keyPair.Lo,
		keyHi:  keyPair.Hi,
	}, nil
}

func computeMaskLo(seed uint32) uint32 {
	return ((^seed ^ loMask1) << 16) | ((seed ^ loMask2) & loMask3)
}

func computeMaskHi(seed uint32) uint32 {
	return ((seed ^ hiMask1) >> 16) | ((^seed ^ hiMask2) & hiMask3)
}

// GetSeed returns the client's seed.
func (c *CryptoService) GetSeed() uint32 {
	return c.seed
}

// GetMasks returns the pair of client masks.
func (c *CryptoService) GetMasks() avatar.KeyPair {
	return avatar.KeyPair{
		Lo: c.maskLo,
		Hi: c.maskHi,
	}
}

// GetKeys returns the pair of version-specific client keys.
func (c *CryptoService) GetKeys() avatar.KeyPair {
	return avatar.KeyPair{
		Lo: c.keyLo,
		Hi: c.keyHi,
	}
}

// Encrypt data using the client's cryptography facilities.
func (c *CryptoService) Encrypt(data []byte) ([]byte, error) {
	// TODO
	return []byte{0x00}, nil
}

// Decrypt data using the client's cryptography facilities.
func (c *CryptoService) Decrypt(data []byte) ([]byte, error) {
	// TODO
	return []byte{0x00}, nil
}

// VerifyLogin verifies that the client's provided login data is valid,
// decrypting it into plaintext account credentials.
func (c *CryptoService) VerifyLogin(src, dest []byte) error {
	// TODO: check client state

	if len(src) != len(dest) {
		return errors.New("buffer size mismatch")
	}

	for i := range src {
		dest[i] = src[i] ^ byte(c.maskLo)

		maskLo := c.maskLo
		maskHi := c.maskHi

		c.maskLo = ((maskLo >> 1) | (maskHi << 31)) ^ c.keyLo
		maskHi = ((maskHi >> 1) | (maskLo << 31)) ^ c.keyHi
		c.maskHi = ((maskHi >> 1) | (maskLo << 31)) ^ c.keyHi
	}

	// Validate dest, ensuring the decrypted next command is a login request
	// and the provided account name and password are NUL-terminated:
	if !(dest[0] == 0x80 && dest[30] == 0x00 && dest[60] == 0x00) {
		return errors.New("unable to decrypt")
	}

	return nil
}
