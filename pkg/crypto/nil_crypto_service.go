package crypto

import (
	"github.com/jadefish/avatar"
)

// NilCryptoService is an avatar.CryptoService that does not modify input when
// encrypting and decrypting data.
type NilCryptoService struct{}

var _ avatar.CryptoService = &NilCryptoService{}

// Encrypt src into dst. The bytes copies to dst are equal to those in src.
func (c NilCryptoService) Encrypt(src []byte, dst []byte) error {
	copy(dst, src)

	return nil
}

// Decrypt src into dst. The bytes copies to dst are equal to those in src.
func (c NilCryptoService) Decrypt(src []byte, dst []byte) error {
	copy(dst, src)

	return nil
}

func (c NilCryptoService) GetSeed() avatar.Seed {
	return avatar.Seed(0)
}
