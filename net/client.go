package net

import (
	"bufio"
	"io"
	"net"

	"github.com/jadefish/avatar"
	"github.com/jadefish/avatar/crypto"
	"github.com/pkg/errors"
)

type clientConn struct {
	io.Reader
	io.Writer
	io.Closer

	conn net.Conn
}

func newClientConn(conn net.Conn) *clientConn {
	return &clientConn{
		Reader: bufio.NewReader(conn),
		Writer: bufio.NewWriter(conn),
		Closer: conn,
		conn:   conn,
	}
}

type Client struct {
	conn    *clientConn
	crypto  avatar.CryptoService
	version avatar.ClientVersion
}

// NewClient creates a new client.
func NewClient(conn net.Conn) *Client {
	return &Client{
		conn:    newClientConn(conn),
		crypto:  crypto.NilCryptoService{},
		version: avatar.ClientVersion{},
	}
}

func (c Client) Disconnect(reason byte) error {
	_, err := c.conn.Write([]byte{avatar.CommandLoginDenied, reason})

	if err != nil {
		return errors.Wrap(err, "disconnect")
	}

	return c.conn.Close()
}

func (c Client) IPAddress() net.IP {
	if seed := c.crypto.GetSeed(); seed != 0 {
		return seed.IPv4()
	}

	return net.ParseIP(c.conn.conn.RemoteAddr().String())
}

func (c *Client) SetCrypto(cs avatar.CryptoService) {
	c.crypto = cs
}

func (c Client) Crypto() avatar.CryptoService {
	return c.crypto
}

func (c *Client) SetVersion(v avatar.ClientVersion) {
	c.version = v
}

func (c Client) Version() avatar.ClientVersion {
	return c.version
}
