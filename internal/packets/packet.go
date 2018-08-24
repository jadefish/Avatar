package packets

// BufferSize represents the maximum packet length. Any packets received that
// are longer than this value are invalid.
const BufferSize = 0xF000

// Packet represents a particular type of client or server packet.
type Packet interface {
	From([]byte) (Packet, error)
	Identifier() uint
	Length() int
	Name() string
}

// BasePacket is a common packet from which all other packets are derived.
type BasePacket struct {
	data []byte
}

// From creates a packet from the provided data.
func (BasePacket) From(buf []byte) (BasePacket, error) {
	panic(errMethodNotImplemented)
}

// Identifier gets the packet's ID.
func (BasePacket) Identifier() uint {
	panic(errMethodNotImplemented)
}

// Length gets the expected packet length.
// If the expected length is 0, the packet is not of a fixed length.
func (BasePacket) Length() int {
	panic(errMethodNotImplemented)
}

// Name gets the human-readable name of the packet.
func (BasePacket) Name() string {
	panic(errMethodNotImplemented)
}
