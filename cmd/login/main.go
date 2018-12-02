package main

import (
	"log"

	"github.com/jadefish/avatar/crypto/bcrypt"
	"github.com/jadefish/avatar/mysql"
	"github.com/jadefish/avatar/net"
)

func main() {
	db, err := mysql.Connect()

	if err != nil {
		log.Fatalln(err)
	}

	as := &mysql.AccountService{db}
	ss := &mysql.ShardService{db}

	// TODO: move cost to separate configuration:
	ps, err := bcrypt.NewPasswordService(bcrypt.DefaultCost)

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
