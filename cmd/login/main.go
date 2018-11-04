package main

import (
	"log"

	"github.com/jadefish/avatar/crypto/bcrypt"
	"github.com/jadefish/avatar/mysql"
	"github.com/jadefish/avatar/net"
	"github.com/pkg/errors"
)

func main() {
	db, err := mysql.Connect()

	if err != nil {
		log.Fatalln(err)
	}

	as := &mysql.AccountService{
		DB: db,
	}

	// TODO: move cost to separate configuration:
	pws, err := bcrypt.NewPasswordService(bcrypt.DefaultCost)

	if err != nil {
		log.Fatalln(err)
	}

	defer server.Stop()

	log.Println("Listening on", server.Address())

	for {
		client, err := server.Accept()

		if err != nil {
			err = errors.Wrap(err, "server accept")
			log.Println(err)

			continue
		}

		go client.Process(*server)
	}
}
