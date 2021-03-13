package binary

import (
	"encoding/binary"
)

// Marshaler is the interface implemented by types that can marshal themselves
// into a binary form.
type Marshaler interface {
	MarshalBinary(order binary.ByteOrder) ([]byte, error)
}
