package sqlite3

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/jadefish/avatar"
)

type accountRepo struct {
	db *sqlx.DB
}

var _ avatar.AccountRepository = &accountRepo{}

// NewAccountRepo creates a new account repository using the provided DB as
// storage.
func NewAccountRepo(db *sqlx.DB) *accountRepo {
	return &accountRepo{db}
}

const createQuery string = `
	insert into accounts
	    (name, email, password_hash, creation_ip, blocked, created_at)
	values (?, ?, ?, ?, ?, ?);
`

// Create an Account, persisting it to the underlying storage.
func (repo *accountRepo) Create(data avatar.CreateAccountData) error {
	err := transaction(repo.db, func(tx *sqlx.Tx) error {
		now := time.Now()
		query := repo.db.Rebind(createQuery)
		_, err := tx.Exec(query,
			data.Name,
			data.Email,
			data.PasswordHash,
			data.CreationIP,
			false,
			now)

		return err
	})

	if err != nil {
		return fmt.Errorf("create account: %w", err)
	}

	return nil
}

const deleteQuery string = `
	update accounts
	set deleted_at = ?,
		updated_at = ?
	where id = ?;
`

// Delete an existing account by ID.
func (repo *accountRepo) Delete(id avatar.EntityID) error {
	err := transaction(repo.db, func(tx *sqlx.Tx) error {
		query := tx.Rebind(deleteQuery)
		now := time.Now()
		_, err := tx.Exec(query, now, now, id)

		return err
	})

	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}

	return nil
}

const getQuery string = `
	select a.*
	from accounts a
	where a.id = ?
	and a.deleted_at is null;
`

// Get retrieves an Account by its ID.
func (repo *accountRepo) Get(id avatar.EntityID) (*avatar.Account, error) {
	account := &avatar.Account{}
	err := transaction(repo.db, func(tx *sqlx.Tx) error {
		query := repo.db.Rebind(getQuery)

		return repo.db.Get(account, query, id)
	})

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

const updateQuery string = `
	update accounts
	set name = ?,
		email = ?,
		password_hash = ?,
		blocked = ?,
		updated_at = ?
	where id = ?;
`

// Update an Account.
func (repo *accountRepo) Update(
	id avatar.EntityID,
	data avatar.UpdateAccountData,
) error {
	err := transaction(repo.db, func(tx *sqlx.Tx) error {
		query := tx.Rebind(updateQuery)
		now := time.Now()
		_, err := tx.Exec(query,
			data.Name,
			data.Email,
			data.PasswordHash,
			data.Blocked,
			now)

		return err
	})

	if err != nil {
		return fmt.Errorf("update account: %w", err)
	}

	return nil
}

const findByNameQuery string = `
	select a.*
	from accounts a
	where a.name = ?
	and a.deleted_at is null
	limit 1;
`

// FindByName retrieves an account by its name.
// No error is returned if the account could not be found.
func (repo *accountRepo) FindByName(name string) (*avatar.Account, error) {
	account := &avatar.Account{}
	err := transaction(repo.db, func(tx *sqlx.Tx) error {
		query := repo.db.Rebind(findByNameQuery)

		return repo.db.Get(account, query, name)
	})

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

const findByEmailQuery string = `
	select a.*
	from accounts a
	where a.email = ?
	and a.deleted_at is null
	limit 1;
`

// FindByEmail retrieves an account by its email.
// No error is returned if the account could not be found.
func (repo *accountRepo) FindByEmail(email string) (*avatar.Account, error) {
	account := &avatar.Account{}
	err := transaction(repo.db, func(tx *sqlx.Tx) error {
		query := repo.db.Rebind(findByEmailQuery)

		return repo.db.Get(account, query, email)
	})

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}
