package login

import (
	"github.com/jadefish/avatar"
	"github.com/pkg/errors"
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
	return ((^seed ^ 0x00001357) << 16) | ((seed ^ 0xffffaaaa) & 0x0000ffff)
}

func computeMaskHi(seed uint32) uint32 {
	return ((seed ^ 0x43210000) >> 16) | ((^seed ^ 0xabcdffff) & 0xffff0000)
}

func (c *CryptoService) GetSeed() uint32 {
	return c.seed
}

func (c *CryptoService) GetMasks() avatar.KeyPair {
	return avatar.KeyPair{
		Lo: c.maskLo,
		Hi: c.maskHi,
	}
}

func (c *CryptoService) GetKeys() avatar.KeyPair {
	return avatar.KeyPair{
		Lo: c.keyLo,
		Hi: c.keyHi,
	}
}

func (c *CryptoService) Encrypt(data []byte) ([]byte, error) {
	// TODO
	return []byte{0x00}, nil
}

func (c *CryptoService) Decrypt(data []byte) ([]byte, error) {
	// TODO
	return []byte{0x00}, nil
}

func (c *CryptoService) VerifyLogin(src, dest []byte) error {
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
