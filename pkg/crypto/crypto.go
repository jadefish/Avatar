package crypto

// CipherName captures the name of a supported password encryption cipher.
type CipherName string

// Valid indicates whether the cipher name is valid.
func (name CipherName) Valid() bool {
	return name == Bcrypt
}

// Supported password ciphers.
const (
	None   CipherName = ""
	Bcrypt CipherName = "bcrypt"
)
