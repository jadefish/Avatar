package crypto

// KeyPair contains a pair of version-specific client encryption keys.
type KeyPair struct {
	Lo, Hi uint32
}

// Encrypter encrypts data.
type Encrypter interface {
	// Encrypt src into dst.
	Encrypt(src []byte, dst []byte) error
}

// Decrypter decrypts data.
type Decrypter interface {
	// Decrypt src into dst.
	Decrypt(src []byte, dst []byte) error
}

// A Cryptographer can both encrypt and decrypt data.
type Cryptographer interface {
	Encrypter
	Decrypter
}

// The NilCryptographer has no behavior.
type NilCryptographer struct{}

// Encrypt does nothing.
func (c NilCryptographer) Encrypt(src []byte, dst []byte) error {
	return nil
}

// Decrypt does nothing.
func (c NilCryptographer) Decrypt(src []byte, dst []byte) error {
	return nil
}

// Seed is a value used to initialize the state of a CryptoService.
type Seed uint32
