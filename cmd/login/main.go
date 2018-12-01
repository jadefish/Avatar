package main

import (
	"log"

	"github.com/jadefish/Avatar/crypto/bcrypt"
	"github.com/jadefish/Avatar/mysql"
	"github.com/jadefish/Avatar/net"
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

	server := net.NewServer(as, pws)
	err = server.Start()

	if err != nil {
		log.Fatalln(err)
	}

	defer server.Stop()
}
