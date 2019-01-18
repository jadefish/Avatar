package packet

import (
	"bytes"
	"errors"
	"io"

	"github.com/jadefish/avatar"
)

var (
	errNoData           = errors.New("no data")
	errInsufficientData = errors.New("insufficient data")
)

// Packet is a formatted unit of data containing command metadata and a
// data payload.
type Packet struct {
	io.Reader

	data   []byte
	length int
	desc   avatar.Descriptor
}

// Data returns a copy of the packet's data.
func (p Packet) Data() []byte {
	buf := make([]byte, len(p.data))
	copy(buf, p.data)

	return buf
}

// Length returns the actual length of the packet's data.
func (p Packet) Length() int {
	return len(p.data)
}

// Descriptor returns the a descriptor containing information about the
// packet.
func (p Packet) Descriptor() avatar.Descriptor {
	return p.desc
}

// New reads data into a Packet.
// Data is read from r until enough has been read to satisfy the expected
// length of the descriptor, fetched via examination of the first byte read.
func New(r io.Reader, cs avatar.CryptoService) (*Packet, error) {
	buf := make([]byte, avatar.BufferSize)

	// Read a single byte (the ID):
	n, err := r.Read(buf[0:1])

	if n < 1 && err == nil {
		err = errNoData
	}

	if err != nil {
		return nil, err
	}

	buf2, err := cs.Decrypt(buf[0:1])

	if err != nil {
		return nil, err
	}

	id := buf2[0]

	// Find a descriptor matching the read ID:
	desc, err := avatar.FindDescriptor(id)

	if err != nil {
		return nil, err
	}

	// Determine expected packet length:
	length := desc.Length()

	if desc.IsVariableLength() {
		buf2, err = cs.Decrypt(buf[1:3])

		if err != nil {
			return nil, err
		}

		length = int(avatar.Encoding.Uint16(buf2[:]))
	}

	// Read, offset by 1 (since we've already read byte 0, the ID):
	n, err = r.Read(buf[1:length])

	if err != nil {
		return nil, err
	}

	if n < 1 {
		return nil, errNoData
	}

	total := n + 1

	// If we haven't yet read enough data to match the expected packet length,
	// keep reading until we do:
	// TODO: limit attempts? set a read timeout?
	if total < length {
		rest := make([]byte, length-total)
		n2, err2 := readRemaining(r, rest)

		if err2 != nil {
			return nil, err2
		}

		offset := total
		total += n2

		if total < length {
			return nil, errInsufficientData
		}

		copy(buf[offset:], rest)
	}

	buf2, err = cs.Decrypt(buf)

	if err != nil {
		return nil, err
	}

	// Make the fat packet struct, keeping only data up to the expected packet
	// length:
	buf = make([]byte, length)
	copy(buf, buf2[:length])

	return makePacket(buf, *desc), nil
}

func makePacket(buf []byte, desc avatar.Descriptor) *Packet {
	return &Packet{
		Reader: bytes.NewReader(buf),
		data:   buf,
		length: len(buf),
		desc:   desc,
	}
}

// readRemaining attempts to read from r until buf is full.
func readRemaining(r io.Reader, buf []byte) (int, error) {
	n := 0

	for n < len(buf) {
		n2, err := r.Read(buf)
		n += n2

		if n < 1 {
			return n, errNoData
		}

		if err != nil {
			return n, err
		}
	}

	return n, nil
}
