package command

import (
	"github.com/jadefish/avatar"
	"github.com/jadefish/avatar/crypto"
)

type LoginSeed struct {
	avatar.BaseCommand

	Seed    avatar.Seed
	Version avatar.ClientVersion
}

// []byte -> cmd
func (cmd *LoginSeed) UnmarshalBinary(data []byte) error {
	cmd.Seed = avatar.Seed(avatar.Encoding.Uint32(data[1:5]))
	cmd.Version.UnmarshalBinary(data[5:])

	return nil
}

func (cmd LoginSeed) Execute(client avatar.Client, server avatar.Server) error {
	// Establish cryptography service:
	cs, err := crypto.NewLoginCryptoService(cmd.Seed, cmd.Version)

	if err != nil {
		return err
	}

	client.SetVersion(cmd.Version)
	client.SetCrypto(cs)

	return nil
}
