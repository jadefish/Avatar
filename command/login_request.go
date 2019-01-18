package command

import (
	"bytes"
	"errors"
	"log"

	"github.com/jadefish/avatar"
)

type LoginRequest struct {
	avatar.BaseCommand

	AccountName     string
	AccountPassword string
	NextLoginKey    uint8
}

// []byte -> cmd
func (cmd *LoginRequest) UnmarshalBinary(data []byte) error {
	cmd.AccountName = string(bytes.Trim(data[1:31], "\000"))
	cmd.AccountPassword = string(bytes.Trim(data[31:61], "\000"))
	cmd.NextLoginKey = data[61]

	return nil
}

func (cmd LoginRequest) Execute(
	client avatar.Client,
	server avatar.Server,
) error {
	accounts := server.AccountService()
	account, err := accounts.GetAccountByName(cmd.AccountName)

	if err != nil {
		return err
	}

	log.Printf("found account: %+v\n", account)

	ps := server.PasswordService()

	if !ps.VerifyPassword(cmd.AccountPassword, account.PasswordHash) {
		return errors.New("unable to authenticate")
	}

	log.Println("passwords match")

	gsl := &GameServerList{}
	b, _ := gsl.MarshalBinary()

	log.Println(b)

	return nil
}
