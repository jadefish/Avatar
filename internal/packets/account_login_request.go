package packets

import "github.com/pkg/errors"

// AccountLoginRequest is the packet sent from clients attempting to
// authenticate against the login server.
type AccountLoginRequest struct {
	BasePacket
	AccountName     string
	AccountPassword string
	NextLoginKey    byte
}

// Create foo
func (AccountLoginRequest) Create(p []byte) (*AccountLoginRequest, error) {
	alr := &AccountLoginRequest{}

	if len(p) != alr.Length() {
		return nil, errors.Wrap(errPacketLengthMismatch, alr.Name())
	}

	alr.data = p
	alr.AccountName = string(p[0:30])
	alr.AccountPassword = string(p[30:61])
	alr.NextLoginKey = p[61]

	return alr, nil
}

// Identifier foo
func (AccountLoginRequest) Identifier() uint {
	return 0x80
}

// Name foo
func (AccountLoginRequest) Name() string {
	return "Account Login Request"
}

// Length foo
func (AccountLoginRequest) Length() int {
	return 1024
}
