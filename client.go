package avatar

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"io"
	"log"
	"net"

	"github.com/pkg/errors"
)

func long2ip(long uint32) string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b[:], long)

	return net.IP(b).String()
}

const (
	BufferSize = 0xF000
)

// ClientVersion represents a client's self-declared version.
type ClientVersion struct {
	Major    uint32
	Minor    uint32
	Patch    uint32
	Revision uint32
}

type State uint8

const (
	StateDisconnected State = iota
	StateConnecting
	StateConnected
	StateNormal
)

// Client represents a connecting or connected user.
type Client struct {
	conn   net.Conn
	reader io.Reader
	writer io.Writer

	State       State
	Version     ClientVersion
	Crypto      Crypto
	AccountName string
	Password    string
}

// New creates a new client for the provided connection.
func New(conn net.Conn) *Client {
	return &Client{
		conn:   conn,
		reader: conn,
		writer: conn,
	}
}

// Disconnect the client.
func (c *Client) Disconnect(reason byte) error {
	_, err := c.Write([]byte{0x53, reason})

	if err != nil {
		return errors.Wrap(err, "disconnect")
	}

	return c.Close()
}

// Close the connection to the client.
func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Write(buf []byte) (int, error) {
	return c.writer.Write(buf)
}

func (c *Client) Read(buf []byte) (int, error) {
	n, err := c.reader.Read(buf)

	if err != nil {
		return n, err
	}

	if n < 1 || n > BufferSize {
		return n, errors.New("bad packet length")
	}

	ip := "?"

	if c.Crypto.Seed > 0 {
		ip = long2ip(c.Crypto.Seed)
	}

	log.Printf(
		"(%s) Read %d/%d bytes:\n%s\n",
		ip,
		n,
		cap(buf),
		hex.Dump(bytes.Trim(buf, "\000")))

	return n, nil
}
