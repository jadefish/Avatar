package avatar

import (
	"io"
	"net"
)

type ClientState uint8

// Client states
const (
	StateDisconnected ClientState = iota
	StateConnecting
	StateAuthenticating
	StateAuthenticated
)

type StateMachine interface {
	GetState() ClientState
	SetState(state ClientState) error
}

type Versionable interface {
	GetVersion() *ClientVersion
	SetVersion(version ClientVersion) error
}

// TODO: break apart
type Client interface {
	io.ReadWriteCloser
	StateMachine
	Versionable
	CryptoService

	RejectLogin(reason LoginRejectionReason) error

	IPAddress() net.IP

	NewCrypto(seed uint32, version ClientVersion) error
	GetCrypto() *CryptoService
}
