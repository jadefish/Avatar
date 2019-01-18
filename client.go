package avatar

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ClientState uint8

// Client states
const (
	StateDisconnected ClientState = iota
	StateConnected
	StateAuthenticating
	StateAuthenticated
	StateLoggedIn
)

// ClientVersion represents a client's self-declared version.
type ClientVersion struct {
	Major    uint32
	Minor    uint32
	Patch    uint32
	Revision uint32
}

func (v ClientVersion) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer

	binary.Write(&b, Encoding, v.Major)
	binary.Write(&b, Encoding, v.Minor)
	binary.Write(&b, Encoding, v.Patch)
	binary.Write(&b, Encoding, v.Revision)

	return b.Bytes(), nil
}

func (v *ClientVersion) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	err := binary.Read(r, Encoding, &v.Major)
	err = binary.Read(r, Encoding, &v.Minor)
	err = binary.Read(r, Encoding, &v.Patch)
	err = binary.Read(r, Encoding, &v.Revision)

	if err != nil {
		return err
	}

	return nil
}

func (v ClientVersion) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Patch, v.Revision)
}

type hasVersion interface {
	Version() ClientVersion
	SetVersion(ClientVersion)
}

type hasCryptoService interface {
	Crypto() CryptoService
	SetCrypto(CryptoService)
}

// Client is a representation of a connected user.
type Client interface {
	hasVersion
	hasCryptoService
}
