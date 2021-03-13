package memory

import (
	"errors"
	"sync"
	"time"

	"github.com/jadefish/avatar"
)

type accountRepo struct {
	m    sync.RWMutex
	id   avatar.EntityID
	data map[avatar.EntityID]*avatar.Account
}

var _ avatar.AccountRepository = &accountRepo{}

var (
	errNoAccount            = errors.New("account does not exist")
	errAccountAlreadyExists = errors.New("account already exists")
)

// NewAccountRepo creates a new account repository using the provided map as
// storage.
func NewAccountRepo(db map[avatar.EntityID]*avatar.Account) *accountRepo {
	return &accountRepo{sync.RWMutex{}, 1, db}
}

// Get a non-deleted account by ID.
// Callers are responsible for acquiring and releasing a read lock.
// If no account is found, a nil pointer is returned.
func (repo *accountRepo) getKept(id avatar.EntityID) *avatar.Account {
	account, ok := repo.data[id]

	if ok && account.DeletedAt == nil {
		return account
	}

	return nil
}

// Get an Account by its ID.
func (repo *accountRepo) Get(id avatar.EntityID) (*avatar.Account, error) {
	repo.m.RLock()
	defer repo.m.RUnlock()

	account := repo.getKept(id)

	if account != nil {
		return account, nil
	}

	return nil, nil
}

// Create a new Account.
func (repo *accountRepo) Create(data avatar.CreateAccountData) error {
	repo.m.Lock()
	defer repo.m.Unlock()

	id := repo.id

	// avoid updating an existing Account:
	if _, ok := repo.data[id]; ok {
		return errAccountAlreadyExists
	}

	now := time.Now()
	account := &avatar.Account{
		ID:             id,
		Name:           data.Name,
		Email:          data.Email,
		PasswordHash:   data.PasswordHash,
		Blocked:        false,
		CreationIP:     data.CreationIP,
		LastLoginIP:    nil,
		LastLoggedInAt: nil,
		CreatedAt:      now,
		UpdatedAt:      &now,
		DeletedAt:      nil,
	}

	repo.data[id] = account
	repo.id += 1

	return nil
}

// Update an Account.
func (repo *accountRepo) Update(id avatar.EntityID, data avatar.UpdateAccountData) error {
	repo.m.Lock()
	defer repo.m.Unlock()

	account := repo.getKept(id)

	if account == nil {
		return errNoAccount
	}

	now := time.Now()

	account.Name = data.Name
	account.Email = data.Email
	account.PasswordHash = data.PasswordHash
	account.Blocked = data.Blocked
	account.LastLoggedInAt = data.LastLoggedInAt
	account.LastLoginIP = data.LastLoginIP
	account.UpdatedAt = &now

	return nil
}

// Delete an existing account by ID.
func (repo *accountRepo) Delete(id avatar.EntityID) error {
	repo.m.Lock()
	defer repo.m.Unlock()

	account := repo.getKept(id)

	if account == nil {
		return errNoAccount
	}

	now := time.Now()
	account.UpdatedAt = &now
	account.DeletedAt = &now

	return nil
}

// FindByName retrieves an account by its name.
func (repo *accountRepo) FindByName(name string) (*avatar.Account, error) {
	repo.m.RLock()
	defer repo.m.RUnlock()

	for _, account := range repo.data {
		if account.Name == name && !account.Deleted() {
			return account, nil
		}
	}

	return nil, nil
}

// FindByEmail retrieves an account by its email.
// No error is returned if the account could not be found.
func (repo *accountRepo) FindByEmail(email string) (*avatar.Account, error) {
	repo.m.RLock()
	defer repo.m.RUnlock()

	for _, account := range repo.data {
		if account.Email == email && !account.Deleted() {
			return account, nil
		}
	}

	return nil, nil
}
