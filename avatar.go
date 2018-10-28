package avatar

import (
	"errors"
	"net"
	"time"
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
	ErrInvalidClientState = errors.New("invalid client state")
)

// BufferSize represents the maximum acceptable length of an incoming packet,
// in bytes
const BufferSize = 0xF000

type Server interface {
	addClient(client Client) error
	FindClient(seed uint32) *Client
	GetClientsByState(state ClientState) ([]*Client, error)
}

type Account struct {
	ID             int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
	LastLoggedInAt *time.Time
	Name           string
	Email          string
	Password       string
	CreationIP     net.IP
	LastLoginIP    *net.IP
}

func (a *Account) IsDeleted() bool {
	return a.DeletedAt != nil
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
