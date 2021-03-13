package binary

import (
	"encoding/binary"
)

// Unmarshaler is the interface implemented by types that can unmarshal a binary
// description of themselves.
type Unmarshaler interface {
	UnmarshalBinary(order binary.ByteOrder, data []byte) error
}
