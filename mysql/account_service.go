package mysql

import (
	"github.com/jmoiron/sqlx"

	av "github.com/jadefish/avatar"
)

// AccountService facilitates interacting with user accounts.
type AccountService struct {
	DB        *sqlx.DB
	Passwords av.PasswordService
}

// GetAccountByID retrieves an account by its ID.
func (s *AccountService) GetAccountByID(id int) (*av.Account, error) {
	var account av.Account

	err := s.DB.Get(&account, `
		SELECT a.*
		FROM accounts a
		WHERE a.id = ?;
	`, id)

	if err != nil {
		return nil, err
	}

	return &account, nil
}

// GetAccountByEmail retrieves an account by its email.
func (s *AccountService) GetAccountByEmail(email string) (*av.Account, error) {
	var account av.Account

	err := s.DB.Get(&account, `
		SELECT a.*
		FROM accounts a
		WHERE a.email= ?;
	`, email)

	if err != nil {
		return nil, err
	}

	return &account, nil
}

// GetAccountByName retrieves an account by its name.
func (s *AccountService) GetAccountByName(name string) (*av.Account, error) {
	var account av.Account

	err := s.DB.Get(&account, `
		SELECT a.*
		FROM accounts a
		WHERE a.name = ?;
	`, name)

	if err != nil {
		return nil, err
	}

	return &account, nil
}
