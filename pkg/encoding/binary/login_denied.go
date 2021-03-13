package binary

import (
	"encoding/binary"
)

type LoginDenied struct {
	Reason LoginDeniedReason
}

var _ SendableCommand = &LoginDenied{}

func (cmd LoginDenied) ID() CommandID {
	return 0x82
}

func (cmd LoginDenied) Name() string {
	return "Login Denied"
}

func (cmd LoginDenied) ExpectedLength() int {
	return 2
}

func (cmd LoginDenied) IsVariableLength() bool {
	return false
}

func (cmd LoginDenied) IsEncrypted() bool {
	return false
}

func (cmd LoginDenied) MarshalBinary(order binary.ByteOrder) ([]byte, error) {
	id := byte(cmd.ID())
	reason := byte(cmd.Reason)

	return []byte{id, reason}, nil
}
