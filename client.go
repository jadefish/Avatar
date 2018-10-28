package avatar

import (
	"net"
)

type ClientState uint8

// Client states
const (
	StateDisconnected ClientState = iota
	StateConnected
	StateAuthenticating
	StateAuthenticated
)

// ClientVersion represents a client's self-declared version.
type ClientVersion struct {
	Major    uint32
	Minor    uint32
	Patch    uint32
	Revision uint32
}

type hasVersion interface {
	Version() *ClientVersion
}

type hasState interface {
	State() ClientState
}

type authenticates interface {
	Authenticate() error
}

type connects interface {
	Connect() error
	Disconnect(reason byte) error
	IPAddress() net.IP
}

type Client interface {
	hasVersion
	hasState
	connects
	authenticates

	GetCrypto() *CryptoService
}
