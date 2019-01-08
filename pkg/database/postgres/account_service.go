package postgres

import (
	"github.com/jmoiron/sqlx"

	"github.com/jadefish/avatar"
)

// AccountService facilitates interacting with user accounts.
type AccountService struct {
	DB *sqlx.DB
}

// GetAccountByID retrieves an account by its ID.
func (s *AccountService) GetAccountByID(id int) (*avatar.Account, error) {
	account := &avatar.Account{}

	err := s.DB.Get(account, `
		SELECT a.*
		FROM accounts a
		WHERE a.id = ?
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
func (s *AccountService) GetAccountByEmail(email string) (*avatar.Account, error) {
	account := &avatar.Account{}

	err := s.DB.Get(account, `
		SELECT a.*
		FROM accounts a
		WHERE a.email= ?
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
func (s *AccountService) GetAccountByName(name string) (*avatar.Account, error) {
	account := &avatar.Account{}

	err := s.DB.Get(account, `
		SELECT a.*
		FROM accounts a
		WHERE a.name = ?
		AND a.deleted_at IS NULL;
	`, name)

	if account.ID == 0 {
		return nil, avatar.ErrNoAccountFound
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}
