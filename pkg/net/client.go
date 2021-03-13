package net

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/jadefish/avatar"
	"github.com/jadefish/avatar/pkg/crypto"
	"github.com/jadefish/avatar/pkg/encoding/binary"
)

// Client retains information about a user connected via a network.
type Client struct {
	crypto avatar.CryptoService

	version avatar.Version
	state   avatar.ClientState

	conn   net.Conn
	reader *bufio.Reader
	buf    []byte
}

// NewClient creates a new client that will use the provided network connection
// for communication.
func NewClient(conn net.Conn) *Client {
	return &Client{
		crypto:  crypto.NilCryptoService{},
		version: avatar.Version{},
		state:   avatar.ClientStateNew,
		conn:    conn,
		reader:  bufio.NewReaderSize(conn, MaxPacketSize),
		buf:     make([]byte, MaxPacketSize),
	}
}

// GetState returns the current state of the client.
func (c Client) GetState() avatar.ClientState {
	return c.state
}

// SetState sets the current state of the client.
func (c *Client) SetState(state avatar.ClientState) {
	c.state = state
}

// Encrypt src into dst using the client's current avatar.CryptoService.
func (c *Client) Encrypt(src []byte, dst []byte) error {
	return c.crypto.Encrypt(src, dst)
}

// Decrypt src into dst using the client's current avatar.CryptoService.
func (c *Client) Decrypt(src []byte, dst []byte) error {
	return c.crypto.Decrypt(src, dst)
}

// SetCryptoService provides the Client with a new avatar.CryptoService to use
// when encrypting and decrypting data.
func (c *Client) SetCryptoService(cs avatar.CryptoService) {
	c.crypto = cs
}

// Disconnect the client, closing its connection.
func (c *Client) Disconnect() error {
	if c.GetState() == avatar.ClientStateDisconnected {
		return nil
	}

	err := c.conn.Close()

	if err == nil {
		c.SetState(avatar.ClientStateDisconnected)
	}

	return err
}

// Identifier provides a unique identifier for the Client.
func (c *Client) Identifier() string {
	return strconv.Itoa(int(c.crypto.GetSeed()))
}

func (c *Client) Addr() net.Addr {
	return c.conn.RemoteAddr()
}

// Read data from Client's connection into p.
// If p is a buffer of length MaxPacketSize, at least three bytes are read.
// Otherwise, if p is a smaller buffer, len(p) bytes are read into p.
func (c *Client) Read(p []byte) (int, error) {
	if len(p) == MaxPacketSize {
		// at least three bytes are required to process variable-length packets.
		return io.ReadAtLeast(c.reader, p, 3)
	} else {
		return io.ReadFull(c.reader, p)
	}
}

// SendCommand a binary.Command to the Client.
func (c *Client) SendCommand(cmd binary.SendableCommand) error {
	data, err := cmd.MarshalBinary(ByteOrder)

	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	if len(data) > MaxPacketSize {
		return errors.New("send: data too large")
	}

	if data[0] != byte(cmd.ID()) {
		return fmt.Errorf("send: unexpected ID 0x%X", data[0])
	}

	n, err := c.conn.Write(data)

	if n < len(data) {
		return errors.New("send: could not send all data")
	}

	dumpBytes("send", c, data, n)

	return err
}

// ReceiveCommand a command from the Client, unmarshalling the data into cmd.
func (c *Client) ReceiveCommand(cmd binary.ReceivableCommand) error {
	var size int

	if cmd.IsVariableLength() {
		size = MaxPacketSize
	} else {
		size = cmd.ExpectedLength()
	}

	buf := make([]byte, size)
	n, err := c.Read(buf)

	if err != nil {
		return fmt.Errorf("receive: %w", err)
	}

	buf = buf[:n]

	if cmd.IsEncrypted() {
		if err := c.crypto.Decrypt(buf, buf); err != nil {
			return fmt.Errorf("receive: %w", err)
		}
	}

	dumpBytes("recv", c, buf, n)

	if buf[0] != byte(cmd.ID()) {
		return fmt.Errorf("receive: unexpected ID 0x%X", buf[0])
	}

	if err = cmd.UnmarshalBinary(ByteOrder, buf); err != nil {
		return fmt.Errorf("receive: %w", err)
	}

	return nil
}

func dumpBytes(msg string, client *Client, p []byte, n int) {
	// TODO: remove/toggle debug logging
	log.Printf(
		"(%s) %s: %d bytes:\n%s\n",
		client.conn.RemoteAddr(),
		msg,
		n,
		hex.Dump(p[0:n]),
	)
}
