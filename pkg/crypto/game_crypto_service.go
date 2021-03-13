package crypto

import (
	"github.com/jadefish/avatar"
)

type GameCryptoService struct {
	seed avatar.Seed
}

var _ avatar.CryptoService = &GameCryptoService{}

func (s GameCryptoService) Encrypt(src []byte, dst []byte) error {
	panic("implement me")
}

func (s GameCryptoService) Decrypt(src []byte, dst []byte) error {
	panic("implement me")
}

func (s GameCryptoService) GetSeed() avatar.Seed {
	return s.seed
}

func NewGameCryptoService(seed avatar.Seed) (*GameCryptoService, error) {
	return &GameCryptoService{
		seed: seed,
	}, nil
}
