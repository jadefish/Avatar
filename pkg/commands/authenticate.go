package commands

import (
	"fmt"

	"github.com/jadefish/avatar"
)

const (
	MaxAccountNameLength = 30
	MaxPasswordLength    = 30
)

// Authenticate a user's credentials.
type Authenticate struct {
	Accounts *avatar.AccountService
}

func (cmd Authenticate) Call(name, password string) (*avatar.Account, error) {
	if len(name) < 1 ||
		len(password) < 1 ||
		len(name) > MaxAccountNameLength ||
		len(password) > MaxPasswordLength {
		return nil, avatar.ErrInvalidCredentials
	}

	account, err := cmd.Accounts.FindByName(name)

	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	_, ok := cmd.Accounts.PasswordService.Verify(password, account.PasswordHash)

	if !ok {
		return nil, avatar.ErrInvalidCredentials
	}

	// TODO: check if account is in use, and return an error if so.

	if account.Blocked {
		return account, avatar.ErrAccountBlocked
	}

	return account, nil
}
