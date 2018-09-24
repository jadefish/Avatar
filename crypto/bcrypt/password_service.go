package bcrypt

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordService is capable of generating and verifying passwords via
// bcrypt.
type PasswordService struct{}

// CreatePassword creates a hash from the provided password.
func (ps *PasswordService) CreatePassword(password []byte) ([]byte, error) {
	p, err := bcrypt.GenerateFromPassword(password, 10)

	if err != nil {
		return nil, err
	}

	return p, nil
}

// ComparePasswords compares the provided password and hash, returning `true`
// if the password matches the hash.
func (ps *PasswordService) ComparePasswords(password, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, password)

	return err == nil
}
