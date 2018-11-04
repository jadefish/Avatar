package bcrypt

import (
	"golang.org/x/crypto/bcrypt"
)

// DefaultCost is the default cost value used when creating new a password
// hash.
const DefaultCost = bcrypt.DefaultCost

type passwordService struct {
	cost int
}

// NewPasswordService creates a new password service capable of generating and
// verifying bcrypt password hashes.
// The cost value controls the number of key expansion operations executed when
// creating a password hash as a power of 2.
func NewPasswordService(cost int) (*passwordService, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return nil, bcrypt.InvalidCostError(cost)
	}

	return &passwordService{
		cost: cost,
	}, nil
}

// CreatePassword creates a bcrypt hash from the provided plaintext password.
func (ps passwordService) CreatePassword(password []byte) ([]byte, error) {
	p, err := bcrypt.GenerateFromPassword(password, ps.cost)

	if err != nil {
		return nil, err
	}

	return p, nil
}

// VerifyPassword compares the provided plaintext password and bcrypt hash.
func (passwordService) VerifyPassword(password, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, password)

	return err == nil
}

// Cost returns the cost value used to create the given bcrypt hash.
func (passwordService) Cost(hash []byte) (int, error) {
	return bcrypt.Cost(hash)
}
