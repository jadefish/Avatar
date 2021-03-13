package avatar

import (
	"errors"
	"net"
	"strings"
	"time"
	"unicode/utf8"
)

const minimumAccountNameLengthInRunes = 4

type Account struct {
	ID             EntityID
	Name           string
	Email          string
	PasswordHash   string
	Blocked        bool
	CreationIP     net.IP
	LastLoginIP    *net.IP
	LastLoggedInAt *time.Time
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time
}

// Deleted indicates whether the Account has been deleted.
func (a Account) Deleted() bool {
	return a.DeletedAt != nil
}

// TODO: change structure to https://stackoverflow.com/a/12326676/7001364 ?

// AccountRepository allows for interacting with an Accounts storage provider.
type AccountRepository interface {
	// Get an Account by its ID.
	Get(EntityID) (*Account, error)

	// Create a new Account.
	Create(CreateAccountData) error

	// Update an existing Account.
	Update(EntityID, UpdateAccountData) error

	// Delete an existing Account.
	Delete(EntityID) error

	// FindByName attempts to find a single Account with the specified name.
	// No error is returned if the Account could not be found.
	FindByName(string) (account *Account, err error)

	// FindByEmail attempts to find a single Account with the specified email.
	// No error is returned if the Account could not be found.
	FindByEmail(string) (account *Account, err error)
}

// AccountService provides facilities for working with Accounts.
type AccountService struct {
	accountRepo     AccountRepository
	PasswordService PasswordService
}

// NewAccountService creates a new service capable of managing Accounts.
func NewAccountService(
	repo AccountRepository,
	passwordService PasswordService,
) *AccountService {
	return &AccountService{
		accountRepo:     repo,
		PasswordService: passwordService,
	}
}

// CreateAccountParameters holds untransformed and unvalidated data for creating
// an Account.
type CreateAccountParameters struct {
	Name       string
	Email      string
	Password   string
	CreationIP net.IP
}

// UpdateAccountParameters holds untransformed and unvalidated data for updating
// an Account.
type UpdateAccountParameters struct {
	Name           string
	Email          string
	Password       string
	LastLoginIP    net.IP
	LastLoggedInAt time.Time
}

// CreateAccountData hold transformed, validated data for creating an Account.
type CreateAccountData struct {
	Name         string
	Email        string
	PasswordHash string
	CreationIP   net.IP
}

// UpdateAccountData holds transformed, validated data for updating an Account.
type UpdateAccountData struct {
	Name           string
	Email          string
	PasswordHash   string
	Blocked        bool
	LastLoginIP    *net.IP
	LastLoggedInAt *time.Time
}

// FromAccount fills the UpdateAccountData struct from an existing Account.
func (d UpdateAccountData) FromAccount(a *Account) {
	d.Name = a.Name
	d.Email = a.Email
	d.PasswordHash = a.PasswordHash
	d.Blocked = a.Blocked
	d.LastLoginIP = a.LastLoginIP
	d.LastLoggedInAt = a.LastLoggedInAt
}

// Account error conditions.
var (
	ErrAccountNotFound       = errors.New("account not found")
	ErrDuplicateAccountName  = errors.New("duplicate account name")
	ErrAccountNameTooShort   = errors.New("account name too short")
	ErrDuplicateAccountEmail = errors.New("email address must be unique")
	ErrAccountAlreadyBlocked = errors.New("account is already blocked")
	ErrAccountNotBlocked     = errors.New("account is not blocked")
)

// Account login error conditions.
var (
	ErrAccountBlocked     = errors.New("account is blocked")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountInUse       = errors.New("account is in use")
)

// Create a new Account.
func (as *AccountService) Create(params CreateAccountParameters) error {
	name := strings.TrimSpace(params.Name)

	// Invariant: account names must be at least 4 characters long.
	if utf8.RuneCountInString(name) < minimumAccountNameLengthInRunes {
		return ErrAccountNameTooShort
	}

	// Invariant: accounts must have a unique name.
	account, err := as.accountRepo.FindByName(name)

	if err != nil {
		return err
	}

	if account != nil {
		return ErrDuplicateAccountName
	}

	email := strings.TrimSpace(params.Email)

	// Invariant: accounts must have a unique email address.
	account, err = as.accountRepo.FindByEmail(email)

	if err != nil {
		return err
	}

	if account != nil {
		return ErrDuplicateAccountEmail
	}

	passwordHash, err := as.PasswordService.Hash(params.Password)

	if err != nil {
		return err
	}

	data := CreateAccountData{
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		CreationIP:   params.CreationIP,
	}

	return as.accountRepo.Create(data)
}

// Get an Account by its ID.
func (as *AccountService) Get(id EntityID) (*Account, error) {
	account, err := as.accountRepo.Get(id)

	if err != nil {
		return nil, err
	}

	if account == nil {
		return nil, ErrAccountNotFound
	}

	return account, nil
}

func (as *AccountService) FindByName(name string) (*Account, error) {
	account, err := as.accountRepo.FindByName(name)

	if err != nil {
		return nil, err
	}

	if account == nil {
		return nil, ErrAccountNotFound
	}

	return account, nil
}

// Block the Account identified by the provided ID.
// If no Account could be found or if the Account is already blocked,
// an error is returned.
func (as *AccountService) Block(id EntityID) error {
	account, err := as.accountRepo.Get(id)

	if err != nil {
		return err
	}

	if account == nil {
		return ErrAccountNotFound
	}

	// Invariant: blocked accounts cannot be blocked.
	if account.Blocked {
		return ErrAccountAlreadyBlocked
	}

	data := UpdateAccountData{}
	data.FromAccount(account)
	data.Blocked = true

	return as.accountRepo.Update(id, data)
}

// Unblock the Account identified by the provided ID.
// If no Account could be found or if the Account is not blocked,
// an error is returned.
func (as *AccountService) Unblock(id EntityID) error {
	account, err := as.accountRepo.Get(id)

	if err != nil {
		return err
	}

	if account == nil {
		return ErrAccountNotFound
	}

	// Invariant: unblocked accounts cannot be unblocked.
	if !account.Blocked {
		return ErrAccountNotBlocked
	}

	data := UpdateAccountData{}
	data.FromAccount(account)
	data.Blocked = false

	return as.accountRepo.Update(id, data)
}
