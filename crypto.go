package avatar

import (
	"encoding/binary"
	"net"
)

// encrypter can encrypt data.
type encrypter interface {
	Encrypt(src []byte) (dest []byte, err error)
}

// decrypter can decrypt data.
type decrypter interface {
	Decrypt(src []byte) (dest []byte, err error)
}

// cryptographer can encrypt and decrypt data.
type cryptographer interface {
	encrypter
	decrypter
}

// KeyPair contains a pair of version-specific client encryption keys.
type KeyPair struct {
	Lo, Hi uint32
}

// CryptoService provides cryptographic services for client data.
type CryptoService interface {
	cryptographer

	GetSeed() Seed
	GetMasks() KeyPair
	GetKeys() KeyPair
}

// Seed is a client encryption seed.
// It is typically the IPv4 address of the client.
type Seed uint32

// ToIPv4 encodes the seed as an IPv4 address.
func (s Seed) ToIPv4() net.IP {
	var b [net.IPv4len]byte
	binary.BigEndian.PutUint32(b[:], uint32(s))

	return net.IP(b[:])
}
