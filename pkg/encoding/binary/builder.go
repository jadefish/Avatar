package binary

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// initialLength controls the initial buffer size for a variable-length binary
// Command.
const initialLength = 256

func makeCommandBytes(id CommandID, length int) []byte {
	var b []byte

	if length == variableLength {
		b = make([]byte, 3, initialLength)

		// Reserve 2 bytes for the command length:
		b[1] = 0x00
		b[2] = 0x00
	} else {
		b = make([]byte, 1, length)
	}

	// The first byte of a binary Command is always its ID.
	b[0] = byte(id)

	return b
}

type builder struct {
	order  binary.ByteOrder
	length int
	err    error
	buf    *bytes.Buffer
}

func write(b *builder, data interface{}) error {
	return binary.Write(b.buf, b.order, data)
}

// NewBuilder creates a new builder capable of constructing a binary
// Command.
func NewBuilder(id CommandID, length int, order binary.ByteOrder) *builder {
	return &builder{
		order:  order,
		length: length,
		err:    nil,
		buf:    bytes.NewBuffer(makeCommandBytes(id, length)),
	}
}

func (b builder) ok() bool {
	return b.err == nil
}

// Err returns the builder's current error, if any.
func (b builder) Err() error {
	return b.err
}

func (b builder) isVariableLength() bool {
	return b.length == variableLength
}

// WriteByte adds a single byte to the builder.
func (b *builder) WriteByte(byte byte) *builder {
	if !b.ok() {
		return b
	}

	b.err = write(b, byte)

	return b
}

// WriteBytes adds the provided bytes to the builder.
func (b *builder) WriteBytes(bytes []byte) *builder {
	if !b.ok() {
		return b
	}

	b.err = write(b, bytes)

	return b
}

// WriteUint8 adds a uint8 value to the builder.
func (b *builder) WriteUint8(u8 uint8) *builder {
	return b.WriteByte(u8)
}

// WriteUint16 adds a uint16 value to the builder.
func (b *builder) WriteUint16(u16 uint16) *builder {
	if !b.ok() {
		return b
	}

	b.err = write(b, u16)

	return b
}

// WriteUint32 adds a uint32 value to the builder.
func (b *builder) WriteUint32(u32 uint32) *builder {
	if !b.ok() {
		return b
	}

	b.err = write(b, u32)

	return b
}

// WriteString adds a fixed-length string to the builder.
func (b *builder) WriteString(string string, length int) *builder {
	if !b.ok() {
		return b
	}

	// TODO: does string encoding need to be considered?

	data := make([]byte, length)
	copy(data, string)

	b.err = write(b, data)

	return b
}

// Bytes finalizes the builder, ensuring it will produce a well-formed command
// and returning the enclosed data or nil and an error when applicable.
func (b builder) Bytes() ([]byte, error) {
	if !b.ok() {
		return nil, b.err
	}

	data := b.buf.Bytes()
	length := len(data)

	if b.isVariableLength() {
		// Write total command length as 2-byte uint16 to bytes [1, 2]:
		b.order.PutUint16(data[1:3], uint16(length))
	} else if b.length != length {
		return nil, fmt.Errorf(
			"malformed command: need %d bytes, but have %d",
			b.length,
			len(data),
		)
	}

	return data, nil
}
