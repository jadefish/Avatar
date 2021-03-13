package binary

import (
	"encoding/binary"
)

type SelectServer struct {
	Index int
}

var _ ReceivableCommand = &SelectServer{}

func (cmd SelectServer) ID() CommandID {
	return 0xA0
}

func (cmd SelectServer) Name() string {
	return "Select Server"
}

func (cmd SelectServer) ExpectedLength() int {
	return 3
}

func (cmd SelectServer) IsVariableLength() bool {
	return false
}

func (cmd SelectServer) IsEncrypted() bool {
	return true
}

func (cmd *SelectServer) UnmarshalBinary(order binary.ByteOrder, data []byte) error {
	cmd.Index = int(order.Uint16(data[1:3]))

	return nil
}
