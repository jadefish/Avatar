package net

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"io"
	"log"
	"net"
	"strings"

	"github.com/jadefish/avatar"
	"github.com/jadefish/avatar/crypto"
	"github.com/jadefish/fizzy"
	"github.com/pkg/errors"
)

type Client struct {
	conn   net.Conn
	crypto avatar.CryptoService

	version *avatar.ClientVersion
	fsm     *fizzy.MooreMachine
}

type authResult struct {
	AccountName string
	Password    []byte
}

var long2ipCache = map[avatar.Seed]net.IP{}

// NewClient sets up a new client.
// After initialization, the client is not yet capable of executing
// cryptographic functions.
func NewClient(conn net.Conn) (*Client, error) {
	fsm := fizzy.NewMachine()
	err := fsm.AddStates([]fizzy.State{
		{"disconnected", avatar.StateDisconnected},
		{"connected", avatar.StateConnected},
		{"authenticating", avatar.StateAuthenticating},
		{"authenticated", avatar.StateAuthenticated},
		{"logged in", avatar.StateLoggedIn},
	})

	if err != nil {
		return nil, errors.Wrap(err, "add states")
	}

	fsm.AddTransitions([]fizzy.Transition{
		{From: "disconnected", To: "disconnected", Input: "disconnect"},
		{From: "disconnected", To: "connected", Input: "connect"},

		{From: "connected", To: "disconnected", Input: "disconnect"},
		{From: "connected", To: "authenticating", Input: "authenticate"},

		{From: "authenticating", To: "disconnected", Input: "disconnect"},
		{From: "authenticating", To: "authenticated", Input: "authenticate"},

		{From: "authenticated", To: "disconnected", Input: "disconnect"},
		{From: "authenticated", To: "logged in", Input: "log in"},
	})

	if err != nil {
		return nil, errors.Wrap(err, "add transitions")
	}

	err = fsm.Start()

	if err != nil {
		return nil, errors.Wrap(err, "start FSM")
	}

	return &Client{
		conn: conn,
		fsm:  fsm,
	}, nil
}

func (c Client) Version() *avatar.ClientVersion {
	return c.version
}

func (c Client) State() avatar.ClientState {
	return c.fsm.Output(nil).(avatar.ClientState)
}

func (c *Client) Connect() error {
	buf, n, err := c.getFirstPayload()

	if err != nil {
		return errors.Wrap(err, "connect")
	}

	if buf[0] != avatar.CommandLoginSeed || n != 21 {
		return errors.New("bad packet")
	}

	// disconnected -> connected
	_, err = c.fsm.Transition("connect")

	if err != nil {
		return errors.Wrap(err, "connect")
	}

	seed := avatar.Seed(binary.BigEndian.Uint32(buf[1:5]))
	c.version = &avatar.ClientVersion{
		Major:    binary.BigEndian.Uint32(buf[5:9]),
		Minor:    binary.BigEndian.Uint32(buf[9:13]),
		Patch:    binary.BigEndian.Uint32(buf[13:17]),
		Revision: binary.BigEndian.Uint32(buf[17:21]),
	}

	ccrs, err := crypto.NewClassicCryptoService(seed, *c.version)

	if err != nil {
		return errors.Wrap(err, "new crypto")
	}

	c.crypto = ccrs

	return err
}

func (c Client) Disconnect(reason byte) error {
	_, err := c.conn.Write([]byte{avatar.CommandDisconnect, reason})

	if err != nil {
		return errors.Wrap(err, "disconnect")
	}

	c.fsm.Transition("disconnect")

	// Server must wait for EOF before the open connection can be closed.

	return err
}

func (c Client) IPAddress() net.IP {
	if c.crypto == nil {
		// Fallback to conn's remote IP:
		return net.ParseIP(c.conn.RemoteAddr().String())
	}

	seed := c.crypto.GetSeed()

	if _, ok := long2ipCache[seed]; !ok {
		long2ipCache[seed] = seed.ToIPv4()
	}

	return long2ipCache[seed]
}

func (c Client) Authenticate() (*authResult, error) {
	buf, n, err := c.read()
	result := &authResult{}

	if err != nil || n < 62 {
		return nil, errors.Wrap(err, "authenticate")
	}

	// connected -> authenticating
	c.fsm.Transition("authenticate")

	dest, err := c.crypto.LoginDecrypt(buf)

	if err != nil {
		return nil, errors.Wrap(err, "login decrypt")
	}

	// Validate dest, ensuring the decrypted next command is a login request
	// and the provided account name and password are NUL-terminated:
	if !(dest[0] == avatar.CommandLoginRequest &&
		dest[30] == 0x00 &&
		dest[60] == 0x00) {
		return nil, errors.New("unable to decrypt")
	}

	result.AccountName = strings.Trim(string(dest[1:31]), "\000")
	result.Password = bytes.Trim(dest[31:61], "\000")

	if len(result.AccountName) < 1 {
		return nil, errors.New("empty account name")
	}

	if len(result.Password) < 1 {
		return nil, errors.New("empty password")
	}

	// authenticating -> authenticated
	c.fsm.Transition("authenticate")

	return result, nil
}

func (c Client) LogIn() error {
	// authenticated -> logged in
	c.fsm.Transition("log in")

	return nil
}

func (c Client) ReceiveShardList(shards []avatar.Shard) error {
	n := len(shards)
	length := 6 + n*40

	buf := make([]byte, 0, length)
	buf = append(buf, avatar.CommandGameServerList)
	buf = append(buf, []byte{0x00, byte(length)}...)
	buf = append(buf, 0xFF) // or: 0xCC, 0x64, 0x5D?
	buf = append(buf, []byte{0x00, byte(n)}...)

	for i, shard := range shards {
		index := i + 1
		buf = append(buf, []byte{0x00, byte(index)}...)

		// server name:
		nb := make([]byte, 32)
		copy(nb[:], []byte(shard.Name))
		buf = append(buf, nb[0:32]...)

		buf = append(buf, byte(shard.PercentFull()))
		buf = append(buf, byte(1))
		buf = append(buf, net.IP.To4(shard.IPAddress)...)
	}

	c.Write(buf)

	return nil
}

func (c Client) read() ([]byte, int, error) {
	return c.readAtMost(avatar.BufferSize)
}

func (c Client) Write(buf []byte) error {
	n, err := c.conn.Write(buf)

	if err != nil {
		return errors.Wrap(err, "client write")
	}

	log.Printf(
		"(%s) Wrote %d bytes:\n%s\n",
		c.IPAddress().String(),
		n,
		hex.Dump(bytes.Trim(buf, "\000")),
	)

	return err
}

func (c Client) readAtMost(size int) ([]byte, int, error) {
	buf := make([]byte, size)
	n, err := c.conn.Read(buf)

	if err == io.EOF {
		// TODO: determine the correct reason value for an EOF disconnect
		return buf, n, c.Disconnect(0x00)
	}

	if err != nil {
		return buf, n, errors.Wrap(err, "read")
	}

	if n < 1 || n > size {
		return buf, n, errors.New("bad packet length")
	}

	log.Printf(
		"(%s) Read %d/%d bytes:\n%s\n",
		c.IPAddress().String(),
		n,
		cap(buf),
		hex.Dump(bytes.Trim(buf, "\000")),
	)

	return buf[:n], n, nil
}

// Retrieve the first required 21 bytes of data.
// Clients seem to occasionally send 1 byte (command), 5 bytes (command +
// seed), or all 21 bytes on first read.
func (c Client) getFirstPayload() ([]byte, int, error) {
	buf, n, err := c.readAtMost(21)

	if err != nil {
		return buf, n, errors.Wrap(err, "get connect payload")
	}

	if n == 21 {
		// All good.
		return buf, n, err
	}

	n2 := 21 - n
	cn := 1
	max := 21 // break after 21 reads of a single byte

	for {
		if n == 21 || cn >= max {
			break
		}

		buf2, n2, err2 := c.readAtMost(n2)

		if err != nil {
			return []byte{}, n2, err2
		}

		buf = append(buf[:n], buf2...)
		n += n2

		cn++
	}

	return buf, n, err
}
