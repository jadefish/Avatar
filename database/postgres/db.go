package postgres

import (
	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Connect to the database.
func Connect() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", "")

	if err != nil {
		return nil, err
	}

	db.MapperFunc(strcase.ToSnake)

	return db, nil
}
