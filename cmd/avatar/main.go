package main

import (
	"encoding/hex"
	"log"
	"net"

	"github.com/jadefish/avatar/internal/crypto"
	p "github.com/jadefish/avatar/internal/packets"
	"github.com/pkg/errors"
)

const (
	defaultHost          = "0.0.0.0"
	defaultPort          = "7775"
	defaultBacklog       = 256
	maxAccountNameLength = 30
)

func getSeed(conn net.Conn) ([4]byte, error) {
	var buf [4]byte
	len, err := conn.Read(buf[:])

	if err != nil {
		return buf, errors.Wrap(err, "read error")
	}

	if len != 4 {
		return buf, errors.Wrap(err, "invalid packet size")
	}

	return buf, nil
}

func handleRequest(conn net.Conn) error {
	log.Println("Client", conn.RemoteAddr().String())

	seed, err := getSeed(conn)

	if err != nil {
		conn.Close()

		return errors.Wrap(err, "seed error")
	}

	log.Printf("seed: %x\n", seed)

	// Signal client to send account info:
	conn.Write([]byte{0x01})

	// Receive account login request packet (of length 62):
	var alr = &p.AccountLoginRequest{}
	buf := make([]byte, alr.Length())
	len, err := conn.Read(buf)

	if err != nil || len < 1 {
		conn.Close()

		return errors.Wrap(err, "C:ALR error")
	}

	alr, err = alr.Create(buf)

	if err != nil {
		conn.Close()

		return errors.Wrap(err, "C:ALR error")
	}

	log.Printf("ALR: %+v\nraw:\n%s\n", alr, hex.Dump(buf))

	return nil
}

func main() {
	crypto.LoadClientKeys()
	l, err := net.Listen("tcp", net.JoinHostPort(defaultHost, defaultPort))

	if err != nil {
		err = errors.Wrap(err, "listen error")
		log.Fatalln(err)
	}

	defer l.Close()

	log.Println("Listening on", l.Addr().String())

	for {
		conn, err := l.Accept()

		if err != nil {
			err = errors.Wrap(err, "accept error")

			continue
		}

		go handleRequest(conn)
	}
}
