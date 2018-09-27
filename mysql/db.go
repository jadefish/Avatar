package mysql

import (
	"github.com/go-sql-driver/mysql"
	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
)

var conf *mysql.Config

func init() {
	conf = mysql.NewConfig()
	conf.User = "root"
	conf.Passwd = "password"
	conf.DBName = "avatar"
	conf.Params = map[string]string{
		"collation": "utf8_unicode_ci",
		"parseTime": "true",
	}
	conf.Collation = "utf8"
}

// Connect to the database.
func Connect() (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", conf.FormatDSN())

	if err != nil {
		return nil, err
	}

	db.MapperFunc(strcase.ToSnake)

	return db, nil
}
