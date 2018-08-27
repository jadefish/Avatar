package crypto

import (
	// "golang.org/x/crypto/twofish"

	"errors"

	"github.com/jadefish/avatar"
)

func LoginDecrypt(c *avatar.Crypto, src, dest []byte) error {
	if len(src) != len(dest) {
		return errors.New("crypto length mismatch")
	}

	for i := range src {
		dest[i] = src[i] ^ byte(c.MaskLo)

		maskLo := c.MaskLo
		maskHi := c.MaskHi

		c.MaskLo = ((maskLo >> 1) | (maskHi << 31)) ^ uint32(c.MasterLo)
		maskHi = ((maskHi >> 1) | (maskLo << 31)) ^ uint32(c.MasterHi)
		c.MaskHi = ((maskHi >> 1) | (maskLo << 31)) ^ uint32(c.MasterHi)
	}

	// Validate dest:
	if !(dest[0] == 0x80 && dest[30] == 0x00 && dest[60] == 0x00) {
		return errors.New("unable to decrypt")
	}

	return nil
}
