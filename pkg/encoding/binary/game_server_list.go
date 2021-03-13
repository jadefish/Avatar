package binary

import (
	"encoding/binary"
	"net"
	"sort"
	"time"
)

type GameServerList struct {
	SystemInfoFlag SystemInfoFlag
	servers        []*GameServerListItem
}

var _ SendableCommand = &GameServerList{}

func (cmd GameServerList) ID() CommandID {
	return 0xA8
}

func (cmd GameServerList) Name() string {
	return "Game Server List"
}

func (cmd GameServerList) ExpectedLength() int {
	return variableLength
}

func (cmd GameServerList) IsVariableLength() bool {
	return true
}

func (cmd GameServerList) IsEncrypted() bool {
	return false
}

func (cmd *GameServerList) AddItem(entry *GameServerListItem) {
	cmd.servers = append(cmd.servers, entry)

	// keep the list of game servers sorted by index:
	sort.Slice(cmd.servers, func(i int, j int) bool {
		return cmd.servers[i].Index < cmd.servers[j].Index
	})
}

func (cmd GameServerList) MarshalBinary(order binary.ByteOrder) ([]byte, error) {
	b := NewBuilder(cmd.ID(), cmd.ExpectedLength(), order).
		WriteByte(byte(cmd.SystemInfoFlag)).
		WriteUint16(uint16(len(cmd.servers)))

	for _, entry := range cmd.servers {
		now := time.Now().UTC()
		_, offset := now.In(entry.TimeZone).Zone()
		seconds := time.Duration(offset) * time.Second
		hours := int8(seconds.Hours())

		b = b.WriteUint16(uint16(entry.Index)).
			WriteString(entry.Name, 32).
			WriteByte(byte(entry.PercentFull)).
			WriteByte(byte(hours)).
			WriteBytes(reverseIP(entry.IP))
	}

	return b.Bytes()
}

func reverseIP(ip []byte) []byte {
	buf := make([]byte, len(ip))

	for i, j := 0, len(ip)-1; i < len(ip); i, j = i+1, j-1 {
		buf[i] = ip[j]
	}

	return buf
}

type GameServerListItem struct {
	Index       int
	Name        string
	PercentFull int
	TimeZone    *time.Location
	IP          net.IP
}
