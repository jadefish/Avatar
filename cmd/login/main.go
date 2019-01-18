package main

import (
	"log"
	"os"
	"strconv"

	"github.com/jadefish/avatar/crypto/bcrypt"
	"github.com/jadefish/avatar/database/postgres"
	"github.com/jadefish/avatar/net"
	"github.com/pkg/errors"

	env "github.com/joho/godotenv"
)

func main() {
	err := env.Load()

	if err != nil {
		log.Fatalln(err)
	}

	db, err := postgres.Connect()

	if err != nil {
		log.Fatalln(errors.Wrap(err, "postgres connect"))
	}

	as := &postgres.AccountService{DB: db}
	ss := &postgres.ShardService{DB: db}

	cost, err := strconv.Atoi(os.Getenv("BCRYPT_COST"))

	if err != nil {
		log.Fatalln(err)
	}

	ps, err := bcrypt.NewPasswordService(cost)

	if err != nil {
		log.Fatalln(err)
	}

	server := net.NewServer(as, ps, ss)
	err = server.Start()

	if err != nil {
		log.Fatalln(err)
	}

	defer server.Stop()
}
