package main

import (
	"encoding/binary"
	"log"
	"net"
	"os"
	"strings"

	"github.com/jadefish/avatar"
	"github.com/jadefish/avatar/internal/crypto"
	"github.com/pkg/errors"
)

const (
	defaultHost     = "0.0.0.0"
	defaultPort     = "7775"
	maxPacketLength = 0xF000
)

type handlerFunc func(*avatar.Client, []byte, int) error

var packetHandlers = map[byte]handlerFunc{
	0xEF: setupClient,
}

func setupClient(c *avatar.Client, buf []byte, n int) error {
	buf2 := make([]byte, 20)

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
		buf2 = buf[1:21]
	}

	// n >= 4
	c.Crypto.Seed = binary.BigEndian.Uint32(buf2[0:4])

	if n >= 20 {
		c.Version = avatar.ClientVersion{
			Major:    binary.BigEndian.Uint32(buf2[4:8]),
			Minor:    binary.BigEndian.Uint32(buf2[8:12]),
			Patch:    binary.BigEndian.Uint32(buf2[12:16]),
			Revision: binary.BigEndian.Uint32(buf2[16:20]),
		}
	}

	log.Println(c.Version)

	// Read first crypto payload:
	buf2 = make([]byte, 62)
	n, err := c.Read(buf2)

	if err != nil {
		return err
	}

	if n != 62 {
		return errors.New("Unexpected crypto packet length")
	}

	// Compute masks:
	c.Crypto.MaskLo = crypto.GetMaskLo(c.Crypto.Seed)
	c.Crypto.MaskHi = crypto.GetMaskHi(c.Crypto.Seed)

	// Get master keys:
	keys, err := crypto.GetClientKeyPair(c.Version)

	if err != nil {
		return err
	}

	c.Crypto.MasterHi, c.Crypto.MasterLo = keys[0], keys[1]

	// Decrypt login credentials:
	buf3 := make([]byte, len(buf2))
	err = crypto.LoginDecrypt(&c.Crypto, buf2, buf3)

	if err != nil {
		return err
	}

	c.AccountName = strings.Trim(string(buf3[1:31]), "\000")
	c.Password = strings.Trim(string(buf3[31:61]), "\000")

	// 0x80: request list of shards
	// 0x91: request list of characters owned by account on shard?
	if !(buf3[0] == 0x80 || buf3[0] == 0x91) {
		return errors.Errorf("Unexpected post-auth packet 0x%x", buf3[0])
	}

	return nil
}

func getHandler(cmd byte) handlerFunc {
	if h, ok := packetHandlers[cmd]; ok {
		return h
	}

	return nil
}

func handle(c *avatar.Client) {
	for {
		buf := make([]byte, avatar.BufferSize)
		n, err := c.Read(buf)

		if err != nil {
			log.Println(err)
			continue
		}

		// Handle commands:
		cmd := buf[0]
		handler := getHandler(cmd)

		if handler == nil {
			err = errors.Errorf("Unknown command %x", cmd)
			log.Println(err)
			continue
		}

		log.Printf("Handling 0x%x...", cmd)
		err = handler(c, buf, n)

		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("done. client: %+v\n", c)
	}
}

func getAddr() string {
	if val, ok := os.LookupEnv("LOGIN_ADDR"); ok {
		if host, port, err := net.SplitHostPort(val); err == nil {
			return net.JoinHostPort(host, port)
		}
	}

	return net.JoinHostPort(defaultHost, defaultPort)
}

func main() {
	err := crypto.LoadClientKeys()

	if err != nil {
		log.Fatalln(err)
	}

	l, err := net.Listen("tcp", getAddr())

	if err != nil {
		log.Fatalln(err)
	}

	defer l.Close()

	log.Println("Listening on", l.Addr().String())

	for {
		conn, err := l.Accept()

		if err != nil {
			err = errors.Wrap(err, "accept error")
			log.Println(err)

			continue
		}

		client := avatar.New(conn)

		go handle(client)
	}
}
