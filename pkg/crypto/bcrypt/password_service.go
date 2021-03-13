package bcrypt

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/jadefish/avatar"
)

// DefaultCost is the default cost value used when creating new a password
// hash.
const DefaultCost = bcrypt.DefaultCost

type passwordService struct {
	cost int
}

var _ avatar.PasswordService = &passwordService{}

// NewPasswordService creates a new password service capable of generating and
// verifying bcrypt password hashes.
// The cost value controls the number of key expansion operations executed when
// creating a password hash as a power of 2.
func NewPasswordService(cost int) (*passwordService, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return nil, bcrypt.InvalidCostError(cost)
	}

	return &passwordService{cost: cost}, nil
}

// Hash creates a bcrypt hash from the provided plaintext password.
func (ps passwordService) Hash(password string) (string, error) {
	p, err := bcrypt.GenerateFromPassword([]byte(password), ps.cost)

	if err != nil {
		return "", fmt.Errorf("bcrypt: create password: %w", err)
	}

	return string(p), nil
}

// Verify compares the provided plaintext password and bcrypt hash.
func (passwordService) Verify(password, hash string) (error, bool) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if err != nil {
		err = fmt.Errorf("bcrypt: verify password: %w", err)
	}

	return err, err == nil
}

// Cost returns the cost value used to create the given bcrypt password hash.
func (passwordService) Cost(hash []byte) (int, error) {
	return bcrypt.Cost(hash)
}

func (ps passwordService) String() string {
	return fmt.Sprintf("bcrypt, cost=%d", ps.cost)
}
