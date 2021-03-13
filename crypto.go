package avatar

// KeyPair contains a pair of version-specific client encryption keys.
type KeyPair struct {
	Lo, Hi uint32
}

// A CryptoService is capable of encrypting and decrypting data.
type CryptoService interface {
	// Encrypt src into dst.
	Encrypt(src []byte, dst []byte) error

	// Decrypt src into dst.
	Decrypt(src []byte, dst []byte) error

	// GetSeed returns the Seed used to initialize the CryptoService.
	GetSeed() Seed
}

// Seed is a value used to initialize the state of a CryptoService.
type Seed uint32

// PasswordService is capable of generating and verifying password hashes.
type PasswordService interface {
	// Hash creates a password hash from the plaintext string.
	Hash(password string) (string, error)

	// Verify that the plaintext password matches the encoded hash.
	Verify(password, hash string) (error, bool)
}
