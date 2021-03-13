package binary

import (
	"encoding/binary"
	"net"
)

type ConnectToGameServer struct {
	IP     net.IP
	Port   int
	NewKey uint32
}

var _ SendableCommand = &ConnectToGameServer{}

func (cmd ConnectToGameServer) ID() CommandID {
	return 0x8C
}

func (cmd ConnectToGameServer) Name() string {
	return "Connect to Game Server"
}

func (cmd ConnectToGameServer) ExpectedLength() int {
	return 11
}

func (cmd ConnectToGameServer) IsVariableLength() bool {
	return false
}

func (cmd ConnectToGameServer) IsEncrypted() bool {
	return false
}

func (cmd ConnectToGameServer) MarshalBinary(order binary.ByteOrder) ([]byte, error) {
	return NewBuilder(cmd.ID(), cmd.ExpectedLength(), order).
		WriteBytes(cmd.IP.To4()).
		WriteUint16(uint16(cmd.Port)).
		WriteUint32(cmd.NewKey).
		Bytes()
}
