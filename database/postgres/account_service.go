package postgres

import (
	"github.com/jmoiron/sqlx"

	"github.com/jadefish/avatar"
)

// accountService facilitates interacting with user accounts.
type accountService struct {
	db *sqlx.DB
}

// NewAccountService creates a new account service backed by PostgreSQL.
func NewAccountService(db *sqlx.DB) *accountService {
	return &accountService{
		db: db,
	}
}

// GetAccountByID retrieves an account by its ID.
func (s *accountService) GetAccountByID(id int) (*avatar.Account, error) {
	account := &avatar.Account{}

	err := s.db.Get(account, `
		SELECT a.*
		FROM accounts a
		WHERE a.id = $1
		AND a.deleted_at IS NULL;
	`, id)

	if account.ID == 0 {
		return nil, avatar.ErrNoAccountFound
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

// GetAccountByEmail retrieves an account by its email.
func (s *accountService) GetAccountByEmail(email string) (*avatar.Account, error) {
	account := &avatar.Account{}

	err := s.db.Get(account, `
		SELECT a.*
		FROM accounts a
		WHERE a.email= $1
		AND a.deleted_at IS NULL;
	`, email)

	if account.ID == 0 {
		return nil, avatar.ErrNoAccountFound
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

// GetAccountByName retrieves an account by its name.
func (s *accountService) GetAccountByName(name string) (*avatar.Account, error) {
	account := &avatar.Account{}

	err := s.db.Get(account, `
		SELECT a.*
		FROM accounts a
		WHERE a.name = $1
		AND a.deleted_at IS NULL;
	`, name)

	if err != nil {
		return nil, err
	}

	if account.ID == 0 {
		return nil, avatar.ErrNoAccountFound
	}

	return account, nil
}
