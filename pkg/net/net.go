package net

import (
	"encoding/binary"
)

// MaxPacketSize indicates the maximum allowed length in bytes of a packet's
// data payload.
const MaxPacketSize = 0xF000

// ByteOrder specifies the default endianness when encoding data for package net
// objects.
var ByteOrder = binary.BigEndian
