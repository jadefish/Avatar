package packets

import (
	"errors"
)

var (
	errMethodNotImplemented = errors.New("Method is not implemented")
	errPacketLengthMismatch = errors.New("Packet length mismatch")
)
