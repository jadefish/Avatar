package avatar

import (
	"errors"
	"net"
	"time"
)

type ClientState uint8

// Client states
const (
	StateDisconnected ClientState = iota
	StateConnecting
	StateAuthenticating
	StateAuthenticated
)

type PopupMessageID byte

// Popup messages
const (
	PopupMessageIncorrectPassword PopupMessageID = iota
	PopupMessageCharacterDoesNotExist
	PopupMessageCharacterAlreadyExists
	PopupMessageGeneric3
	PopupMessageGeneric4
	PopupMessageAnotherCharacterIsLoggedIn
	PopupMessageSynchronizationError
	PopupMessageIdleTimeout
	PopupMessageGeneric8
	PopupMessageCharacterTransfer
)

type LoginRejectionReason byte

// Login rejection reasons
const (
	LoginRejectionInvalidAccount LoginRejectionReason = iota
	LoginRejectionAccountInUse
	LoginRejectionAccountBlocked
	LoginRejectionInvalidPassword
	LoginRejectionCommunicationProblem
	LoginRejectionIGRConcurrencyLimitMet
	LoginRejectionIGRTimeLimitMet
	LoginRejectionIGRGeneralAuthFailure
)

var (
	ErrNoAccountFound     = errors.New("no account found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// BufferSize represents the maximum acceptable length of an incoming packet,
// in bytes
const BufferSize = 0xF000

type Server interface {
	AddClient(Client) error
}

type Client interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
	RejectLogin(reason LoginRejectionReason) error
	GetVersion() ClientVersion
	GetState() ClientState
	SetState(state ClientState) error
	IPAddress() net.IP
}

type CryptoService interface {
	GetSeed() uint32
	GetMasks() KeyPair
	GetKeys() KeyPair
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

type KeyPair struct {
	Lo, Hi uint32
}

// ClientVersion represents a client's self-declared version.
type ClientVersion struct {
	Major    uint32
	Minor    uint32
	Patch    uint32
	Revision uint32
}

type Account struct {
	ID             int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastLoggedInAt *time.Time
	Name           string
	Email          string
	Password       string
	CreationIP     net.IP
	LastLoginIP    *net.IP
}

type AccountService interface {
	GetAccountByID(id int) (*Account, error)
	GetAccountByEmail(email string) (*Account, error)
	GetAccountByName(name string) (*Account, error)
}

// PasswordService is capable of generating and verifying password hashes.
type PasswordService interface {
	CreatePassword(password []byte) ([]byte, error)
	ComparePasswords(password, hash []byte) bool
}
