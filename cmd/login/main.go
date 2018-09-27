package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"log"
	"strings"

	"github.com/jadefish/avatar"
	"github.com/jadefish/avatar/crypto/bcrypt"
	"github.com/jadefish/avatar/internal/app/login"
	"github.com/jadefish/avatar/mysql"
	"github.com/jadefish/avatar/net"
	"github.com/pkg/errors"
)

var (
	errInvalidCredentials = errors.New("invalid credentials")
)

type handlerFunc func(*net.Client, []byte, int) error

var packetHandlers = map[byte]handlerFunc{
	0xEF: setupClient,
	0x80: loginRequest,
}

func setupClient(c *net.Client, buf []byte, n int) error {
	buf2 := make([]byte, 20)

	// Some clients send 21 bytes all at once, while others send a first byte
	// followed by the remaining 20 bytes.
	// Since the first byte is always 0xEF, discard it and read or copy the
	// remaining 20 bytes:
	if n == 1 {
		// Read remaining 20 bytes:
		n2, err := c.Read(buf2)
		n = n2

		if err != nil {
			return err
		}

		if n < 4 {
			return errors.New("not enough data")
		}
	} else if n == 21 {
		copy(buf2[:], buf[1:21])
	}

	// n >= 4
	seed := binary.BigEndian.Uint32(buf2[0:4])
	version := &avatar.ClientVersion{}

	if n >= 20 {
		// Parse verison:
		version.Major = binary.BigEndian.Uint32(buf2[4:8])
		version.Minor = binary.BigEndian.Uint32(buf2[8:12])
		version.Patch = binary.BigEndian.Uint32(buf2[12:16])
		version.Revision = binary.BigEndian.Uint32(buf2[16:20])
	}

	// Read first crypto payload:
	cSrc := make([]byte, 62)
	n, err := c.Read(cSrc)

	if err != nil {
		return err
	}

	if n != 62 {
		return errors.New("Unexpected crypto packet length")
	}

	// Set up client's cryptography service:
	crypto, err := login.NewCrypto(seed, version)

	if err != nil {
		return err
	}

	// Decrypt login credentials:
	cDest := make([]byte, n)
	err = crypto.VerifyLogin(cSrc, cDest)

	log.Printf("dest:\n%s\n", hex.Dump(cDest))

	if err != nil {
		return err
	}

	accountName := strings.Trim(string(cDest[1:31]), "\000")
	db, err := mysql.Connect()

	if err != nil {
		return err
	}

	defer db.Close()

	ps := &bcrypt.PasswordService{}
	as := &mysql.AccountService{
		DB:        db,
		Passwords: ps,
	}
	account, err := as.GetAccountByName(accountName)

	if err != nil {
		return err
	}

	log.Printf("found account: %+v\n", account)

	// Ensure provided password matches retrieved account's password:
	password := bytes.Trim(cDest[31:61], "\000")

	if !as.Passwords.ComparePasswords(password, []byte(account.Password)) {
		c.Disconnect(avatar.DisconnectReasonIncorrectPassword)

		return errInvalidCredentials
	}

	// 0x80: request list of shards
	// 0x91: request list of characters owned by account on shard?
	if !(cDest[0] == 0x80 || cDest[0] == 0x91) {
		return errors.Errorf("Unexpected post-auth packet 0x%x", cDest[0])
	}

	return nil
}

func loginRequest(c *net.Client, buf []byte, n int) error {
	// TODO
	return nil
}

func getHandler(cmd byte) handlerFunc {
	if h, ok := packetHandlers[cmd]; ok {
		return h
	}

	return nil
}

func handle(c *net.Client) {
	for {
		buf := make([]byte, avatar.BufferSize)
		n, err := c.Read(buf)

		if err != nil {
			log.Println(err)
			return
		}

		// Handle commands:
		cmd := buf[0]
		handler := getHandler(cmd)

		if handler == nil {
			err = errors.Errorf("Unknown command 0x%x", cmd)
			log.Println(err)
			continue
		}

		log.Printf("Handling 0x%x...", cmd)
		err = handler(c, buf, n)

		if err != nil {
			log.Println(err)
			continue
		}

		// pw, err := bcrypt.GenerateFromPassword([]byte(c.Password), 10)
		// log.Println("brcypt:", string(pw))
		//
		// log.Printf("done. client: %+v\n", c)
		// err = c.Authenticate()
		//
		// if err != nil {
		// 	log.Println(err)
		// 	return
		// } else {
		// 	log.Println("auth'd successfully")
		// }
	}
}

func main() {
	server := net.NewServer()
	err := server.Start()

	if err != nil {
		log.Fatalln(err)
	}

	defer server.Stop()

	log.Println("Listening on", server.Address())

	for {
		conn, err := server.Accept()

		if err != nil {
			err = errors.Wrap(err, "accept error")
			log.Println(err)

			continue
		}

		client := net.NewClient(conn)

		go handle(client)
	}
}
