package avatar

import (
	"fmt"
)

const variableLength = -1

// Descriptor contains information about a specific packet or a command
// constructed from a packet.
type Descriptor struct {
	id        byte
	name      string
	length    int
	supported bool
}

// ID retrieves a unique byte identitier for the packet.
func (d Descriptor) ID() byte {
	return d.id
}

// Name retireves a friendly string identitifer for the packet
func (d Descriptor) Name() string {
	return d.name
}

// Length retrieves the expected length of the packet.
// If the length varies, -1 is returned.
func (d Descriptor) Length() int {
	return d.length
}

// IsVariableLength determines whether the expected length of the packet is
// fixed.
// If the packet is variable length, the length is stored as a 16-bit unsigned
// integer in the second and third byte.
func (d Descriptor) IsVariableLength() bool {
	return d.length == variableLength
}

// IsSupported determines whether the server will attempt to process the
// packet.
func (d Descriptor) IsSupported() bool {
	return d.supported
}

var descriptors = map[byte]Descriptor{
	0x80: {0x80, "Login request", 62, true},
	0xA8: {0xA8, "Game server list", -1, true},
	0xEF: {0xEF, "Login seed", 21, true},
}

// FindDescriptor looks up a descriptor by ID.
func FindDescriptor(id byte) (*Descriptor, error) {
	p, ok := descriptors[id]

	if !ok {
		return nil, unknownCommandError{id}
	}

	if !p.supported {
		return nil, unsupportedCommandError{id}
	}

	return &p, nil
}

const formatString = "%s command 0x%X"

type unknownCommandError struct {
	id byte
}

func (e unknownCommandError) Error() string {
	return fmt.Sprintf(formatString, "unknown", e.id)
}

type unsupportedCommandError struct {
	id byte
}

func (e unsupportedCommandError) Error() string {
	return fmt.Sprintf(formatString, "unsupported", e.id)
}
