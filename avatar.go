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

// Record contains common fields for database records.
type Record struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// IsDeleted determines whether the record has been deleted.
func (r Record) IsDeleted() bool {
	return r.DeletedAt != nil
}

type Account struct {
	Record

	LastLoggedInAt *time.Time
	Name           string
	Email          string
	PasswordHash   string
	CreationIP     net.IP
	LastLoginIP    *net.IP
}

// AccountService provides methods for working with user accounts.
type AccountService interface {
	GetAccountByID(id int) (*Account, error)
	GetAccountByEmail(email string) (*Account, error)
	GetAccountByName(name string) (*Account, error)
}

// PasswordService is capable of generating and verifying password hashes.
type PasswordService interface {
	CreatePassword(password []byte) ([]byte, error)
	VerifyPassword(password, hash []byte) bool
}

// Shard contains information about a game server.
type Shard struct {
	Name        string
	PercentFull int
	TimeZone    int
	IPAddress   net.IP
}
