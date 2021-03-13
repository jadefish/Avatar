package crypto

import (
	"errors"
	"fmt"

	"github.com/jadefish/avatar"
)

// LoginCryptoService is an avatar.CryptoService compatible with modern
// "classic" (2D) capable of decrypting data during the login process.
// LoginCryptoService cannot encrypt data.
//
// LoginCryptoService implements a rolling cipher on both ends. Decrypting data
// modifies the internal state of the service and affects subsequent decryption
// operations.
type LoginCryptoService struct {
	seed           avatar.Seed
	maskLo, maskHi uint32 // mutable
	keyLo, keyHi   uint32 // "master" lo and hi
}

var _ avatar.CryptoService = &LoginCryptoService{}

// NewLoginCryptoService creates a new LoginCryptoService.
func NewLoginCryptoService(
	seed avatar.Seed,
	version *avatar.Version,
) (*LoginCryptoService, error) {
	keyPair, err := GetKeyPair(version)

	if err != nil {
		return nil, fmt.Errorf("create login crypto: %w", err)
	}

	const (
		loMask1 = 0x00001357
		loMask2 = 0xffffaaaa
		loMask3 = 0x0000ffff
		hiMask1 = 0x43210000
		hiMask2 = 0xabcdffff
		hiMask3 = 0xffff0000
	)

	maskLo := uint32(((^seed ^ loMask1) << 16) | ((seed ^ loMask2) & loMask3))
	maskHi := uint32(((seed ^ hiMask1) >> 16) | ((^seed ^ hiMask2) & hiMask3))

	return &LoginCryptoService{
		seed:   seed,
		maskLo: maskLo,
		maskHi: maskHi,
		keyLo:  keyPair.Lo,
		keyHi:  keyPair.Hi,
	}, nil
}

// Encrypt is an invalid operation for LoginCryptoService and always panics.
func (cs *LoginCryptoService) Encrypt(src []byte, dst []byte) error {
	panic(errors.New("login crypto service cannot encrypt data"))
}

// Decrypt src into dst. Bytes in dst are overridden as decryption occurs.
//
// If len(src) > len(dst), up to len(dst) bytes are decrypted and an error is
// returned.
func (cs *LoginCryptoService) Decrypt(src []byte, dst []byte) error {
	for i := 0; i < len(src); i++ {
		if i >= len(dst) {
			return errors.New("decrypt: exceeded output buffer length")
		}

		dst[i] = src[i] ^ byte(cs.maskLo)

		maskLo := cs.maskLo
		maskHi := cs.maskHi

		// The login crypto service implements a rolling cipher on both ends,
		// so cs's masks are mutated whenever data is decrypted.
		cs.maskLo = ((maskLo >> 1) | (maskHi << 31)) ^ cs.keyLo
		maskHi = ((maskHi >> 1) | (maskLo << 31)) ^ cs.keyHi
		cs.maskHi = ((maskHi >> 1) | (maskLo << 31)) ^ cs.keyHi
	}

	return nil
}

func (cs LoginCryptoService) GetSeed() avatar.Seed {
	return cs.seed
}
