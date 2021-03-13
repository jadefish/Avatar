package avatar_test

import (
	"net"
	"testing"

	"github.com/jadefish/avatar/pkg/crypto/bcrypt"

	. "github.com/jadefish/avatar"
	"github.com/jadefish/avatar/pkg/storage/memory"
)

func accountRepo(t *testing.T) AccountRepository {
	t.Helper()

	db := make(map[EntityID]*Account)

	return memory.NewAccountRepo(db)
}

func passwordService(t *testing.T) PasswordService {
	t.Helper()

	ps, err := bcrypt.NewPasswordService(bcrypt.DefaultCost)

	if err != nil {
		t.Fatal(err)
	}

	return ps
}

func TestAccountService_Create(t *testing.T) {
	ar := accountRepo(t)
	ps := passwordService(t)
	as := NewAccountService(ar, ps)

	// Invariant: account names must be at least 4 characters long.
	params := CreateAccountParameters{Name: "yo"}
	err := as.Create(params)

	switch err {
	case ErrAccountNameTooShort:
	case nil:
		t.Error("failed invariant: account names must be >= 4 runes long")
	default:
		t.Fatal("unable to create account")
	}

	// Invariant: accounts must have a unique name.
	name := "duplicate_account_name"
	email := "what@ever.com"
	err = as.Create(CreateAccountParameters{
		Name:     name,
		Email:    email,
		Password: "meowmix",
	})

	if err != nil {
		t.Fatal("unable to create account")
	}

	err = as.Create(CreateAccountParameters{
		Name:     name,
		Email:    "another@email.com",
		Password: "woofwoof",
	})

	switch err {
	case ErrDuplicateAccountName:
	case nil:
		t.Error("failed invariant: Accounts must have a unique name")
	default:
		t.Fatal("unable to create Account")
	}

	// Invariant: accounts must have a unique email address.
	err = as.Create(CreateAccountParameters{
		Name:     "another_account",
		Email:    email,
		Password: "hunter123",
	})

	switch err {
	case ErrDuplicateAccountEmail:
	case nil:
		t.Error("failed invariant: Accounts must have a unique email address")
	default:
		t.Fatal("unable to create Account")
	}

	// Attempted creation of an Account from valid parameters should not return
	// an error.
	err = as.Create(CreateAccountParameters{
		Name:       "my well formed account name",
		Email:      "a.super.cool.email@a_domain.com",
		Password:   "unguessable",
		CreationIP: net.IPv4zero,
	})

	if err != nil {
		t.Error("create Account with valid parameters failed")
	}
}

func TestAccountService_Block(t *testing.T) {
	ar := accountRepo(t)
	ps := passwordService(t)
	as := NewAccountService(ar, ps)

	name := "account'"
	err := as.Create(CreateAccountParameters{
		Name:       name,
		Email:      "email@address.com",
		Password:   "unguessable",
		CreationIP: net.IPv4zero,
	})

	if err != nil {
		t.Fatal("unable to create account")
	}

	account, err := ar.FindByName(name)

	if err != nil || account == nil {
		t.Fatal("unable to retrieve account")
	}

	id := account.ID
	err = as.Block(id)

	if err != nil {
		t.Error("failed to block valid account")
	}

	// Invariant: blocked accounts cannot be blocked.
	err = as.Block(id)

	if err != ErrAccountAlreadyBlocked {
		t.Error("failed invariant: blocked accounts cannot be blocked")
	}
}

func TestAccountService_Unblock(t *testing.T) {
	ar := accountRepo(t)
	ps := passwordService(t)
	as := NewAccountService(ar, ps)

	name := "account'"
	err := as.Create(CreateAccountParameters{
		Name:       name,
		Email:      "email@address.com",
		Password:   "unguessable",
		CreationIP: net.IPv4zero,
	})

	if err != nil {
		t.Fatal("unable to create account")
	}

	account, err := ar.FindByName(name)

	if err != nil || account == nil {
		t.Fatal("unable to retrieve account")
	}

	id := account.ID

	// Invariant: unblocked accounts cannot be unblocked.
	err = as.Unblock(id)

	if err != ErrAccountNotBlocked {
		t.Error("failed invariant: unblocked accounts cannot be unblocked")
	}

	_ = as.Block(id)
	err = as.Unblock(id)

	if err != nil {
		t.Error("failed to unblock valid account")
	}
}
