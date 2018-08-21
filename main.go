package main

import (
	"log"
	"net"

	"github.com/pkg/errors"
)

const (
	defaultHost          = "0.0.0.0"
	defaultPort          = "7775"
	defaultBacklog       = 256
	maxAccountNameLength = 30
)

var (
	errPacketLengthMismatch = errors.New("Packet length mismatch")
)

type packet interface {
	Identifier() uint
	Name() string
	FromPacket([]byte) packet
	Length() int
}

type accountLoginRequest struct {
	AccountName     string
	AccountPassword string
	NextLoginKey    byte
}

func (accountLoginRequest) Identifier() uint {
	return 0x80
}

func (accountLoginRequest) Name() string {
	return "Account Login Request"
}

func (accountLoginRequest) Length() int {
	return 62
}

func (accountLoginRequest) FromPacket(p []byte) (*accountLoginRequest, error) {
	alr := &accountLoginRequest{}

	if len(p) != alr.Length() {
		return nil, errors.Wrap(errPacketLengthMismatch, alr.Name())
	}

	alr.AccountName = string(p[1:31])
	alr.AccountPassword = string(p[31:61])
	alr.NextLoginKey = p[61]

	return alr, nil
}

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
		return errors.Wrap(err, "seed error")
	}

	log.Printf("seed: %x\n", seed)

	// Signal client to send account info:
	conn.Write([]byte{0x00})

	// Receive account login request packet (of length 62):
	var alr = &accountLoginRequest{}
	buf := make([]byte, alr.Length())
	len, err := conn.Read(buf)

	if err != nil || len < 1 {
		return errors.Wrap(err, "C:ALR error")
	}

	alr, err = alr.FromPacket(buf)

	if err != nil {
		return errors.Wrap(err, "C:ALR error")
	}

	log.Printf("ALR: %+v\n", alr)

	return nil
}

func main() {
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
