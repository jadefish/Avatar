package avatar

import (
	"encoding/binary"
	"net"
)

// encrypter can encrypt data.
type encrypter interface {
	Encrypt(src []byte) (dest []byte, err error)
}

// Decrypter can decrypt data.
type Decrypter interface {
	Decrypt(src []byte) (dest []byte, err error)
}

// Cryptographer is capable of encrypting and decrypting data.
type Cryptographer interface {
	Encrypter
	Decrypter
}

type CryptoService interface {
	Cryptographer

	GetSeed() uint32
	GetMasks() KeyPair
	GetKeys() KeyPair
	VerifyLogin(src, dest []byte) error
}

type Seed uint32

// ToIPv4 encodes the seed as an IPv4 address.
func (s Seed) ToIPv4() net.IP {
	b := make([]byte, net.IPv4len)
	binary.BigEndian.PutUint32(b[:], uint32(s))

	return net.IP(b)
}

type KeyPair struct {
	Lo, Hi uint32
}
