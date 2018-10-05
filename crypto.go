package avatar

// Encrypter can encrypt data.
type Encrypter interface {
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

type KeyPair struct {
	Lo, Hi uint32
}
