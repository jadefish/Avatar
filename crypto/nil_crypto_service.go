package crypto

import (
	"github.com/jadefish/avatar"
)

// NilCryptoService is a cryptography service which returns its input.
type NilCryptoService struct{}

// GetSeed returns the client's seed.
func (c NilCryptoService) GetSeed() avatar.Seed {
	return 0
}

// GetMasks returns the pair of client masks.
func (c NilCryptoService) GetMasks() avatar.KeyPair {
	return avatar.KeyPair{
		Lo: 0,
		Hi: 0,
	}
}

// GetKeys returns the pair of version-specific client keys.
func (c NilCryptoService) GetKeys() avatar.KeyPair {
	return avatar.KeyPair{
		Lo: 0,
		Hi: 0,
	}
}

// Encrypt encrypts src, returning the encrypted data.
func (c NilCryptoService) Encrypt(src []byte) ([]byte, error) {
	return src, nil
}

// Decrypt src, returning the unencrypted data.
func (c NilCryptoService) Decrypt(src []byte) ([]byte, error) {
	return src, nil
}
