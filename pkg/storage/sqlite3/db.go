package sqlite3

import (
	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"
)

// Connect to a SQLite database using the provided connection string.
func Connect(connection string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", connection)

	if err != nil {
		return nil, err
	}

	db.MapperFunc(strcase.ToSnake)

	return db, nil
}

// execute the provided function in a new transaction.
// if fn returns an error or panics, the transaction is rolled back and the
// offending error is returned. otherwise, the transaction is committed.
func transaction(db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
	// from https://stackoverflow.com/a/23502629/7001364
	tx, err := db.Beginx()

	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)

	return err
}
