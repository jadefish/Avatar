package login

import (
	"errors"
	"fmt"

	"github.com/jadefish/avatar"
	"github.com/jadefish/avatar/pkg/crypto"
)

// Decrypter is capable of decrypting data during the login process.
//
// Decrypter implements a rolling cipher on both ends. Decryption operations
// modify the internal state of the service and affect subsequent operations.
type Decrypter struct {
	seed           crypto.Seed
	maskLo, maskHi uint32 // mutable
	keyLo, keyHi   uint32 // "master" lo and hi
}

// NewDecrypter initializes a new crypto service using the provided seed and
// version.
func NewDecrypter(seed crypto.Seed, version *avatar.Version) (*Decrypter, error) {
	keyPair, err := crypto.GetKeyPair(version)

	if err != nil {
		return nil, fmt.Errorf("create login crypto: %w", err)
	}

	const (
		loMask1 = 0x00001357
		loMask2 = 0xffffaaaa
		loMask3 = 0x0000ffff
		hiMask1 = 0x43210000
		hiMask2 = 0xabcdffff
		hiMask3 = 0xffff0000
	)

	maskLo := uint32(((^seed ^ loMask1) << 16) | ((seed ^ loMask2) & loMask3))
	maskHi := uint32(((seed ^ hiMask1) >> 16) | ((^seed ^ hiMask2) & hiMask3))

	return &Decrypter{
		seed:   seed,
		maskLo: maskLo,
		maskHi: maskHi,
		keyLo:  keyPair.Lo,
		keyHi:  keyPair.Hi,
	}, nil
}

// Decrypt src into dst. Bytes in dst are overridden as decryption occurs.
//
// If len(src) > len(dst), up to len(dst) bytes are decrypted and an error is
// returned.
//
// Note that the Decrypter's internal state is mutated during decryption.
func (d *Decrypter) Decrypt(src []byte, dst []byte) error {
	for i := 0; i < len(src); i++ {
		if i >= len(dst) {
			return errors.New("decrypt: exceeded output buffer length")
		}

		dst[i] = src[i] ^ byte(d.maskLo)

		maskLo := d.maskLo
		maskHi := d.maskHi

		// NB: the masks are mutated as data is decrypted.
		d.maskLo = ((maskLo >> 1) | (maskHi << 31)) ^ d.keyLo
		maskHi = ((maskHi >> 1) | (maskLo << 31)) ^ d.keyHi
		d.maskHi = ((maskHi >> 1) | (maskLo << 31)) ^ d.keyHi
	}

	return nil
}
