package net

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"log"
	"net"

	"github.com/jadefish/avatar"
	"github.com/jadefish/avatar/internal/app/login"
	"github.com/pkg/errors"
)

// Client represents a connected user.
type Client struct {
	conn    net.Conn
	state   avatar.ClientState
	version avatar.ClientVersion
	account *avatar.Account
	crypto  *login.CryptoService
}

// NewClient creates a new client for the provided connection.
func NewClient(conn net.Conn) *Client {
	client := &Client{
		conn:  conn,
		state: avatar.StateDisconnected,
	}

	return client
}

// Read from the client.
func (c *Client) Read(buf []byte) (int, error) {
	n, err := c.conn.Read(buf)

	if err != nil {
		return n, err
	}

	if n < 1 || n > avatar.BufferSize {
		return n, errors.New("bad packet length")
	}

	ip := c.IPAddress()

	log.Printf(
		"(%s) Read %d/%d bytes:\n%s\n",
		ip,
		n,
		cap(buf),
		hex.Dump(bytes.Trim(buf, "\000")))

	return n, nil
}

// Write to the client.
func (c *Client) Write(buf []byte) (int, error) {
	return c.conn.Write(buf)
}

// Close the client's connection.
func (c *Client) Close() error {
	err := c.conn.Close()

	if err == nil {
		c.SetState(avatar.StateDisconnected)
	}

	return errors.Wrap(err, "close")
}

// RejectLogin terminates the client's current authentication attempt.
func (c *Client) RejectLogin(reason avatar.LoginRejectionReason) error {
	// It is invalid to reject a login process for a client that has already
	// logged in.
	if c.GetState() > avatar.StateAuthenticating {
		return errors.Wrap(avatar.ErrInvalidClientState, "cannot reject login")
	}

	_, err := c.conn.Write([]byte{0x82, byte(reason)})

	if err == nil {
		c.SetState(avatar.StateDisconnected)
	}

	return errors.Wrap(err, "disconnect")
}

// GetVersion retrieve's the client's self-reported version.
func (c *Client) GetVersion() avatar.ClientVersion {
	return c.version
}

// GetState retireves the current state of the client.
func (c *Client) GetState() avatar.ClientState {
	return c.state
}

// SetState transitions the client into a new state.
func (c *Client) SetState(state avatar.ClientState) error {
	// TODO: validate transition
	c.state = state

	return nil
}

// IPAddress returns the IP address of the client.
func (c *Client) IPAddress() string {
	ip := "unknown"

	if seed := c.crypto.GetSeed(); seed > 0 {
		ip = long2ip(seed).String()
	}

	return ip
}

func long2ip(long uint32) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b[:], long)

	return net.IP(b)
}
