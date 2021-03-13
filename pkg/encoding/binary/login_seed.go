package binary

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/jadefish/avatar"
)

type LoginSeed struct {
	Seed    avatar.Seed
	Version avatar.Version
}

var _ ReceivableCommand = &LoginSeed{}

func (cmd LoginSeed) ID() CommandID {
	return 0xEF
}

func (cmd LoginSeed) Name() string {
	return "Login Seed"
}

func (cmd LoginSeed) ExpectedLength() int {
	return 21
}

func (cmd LoginSeed) IsVariableLength() bool {
	return false
}

func (cmd LoginSeed) IsEncrypted() bool {
	return false
}

func (cmd *LoginSeed) UnmarshalBinary(order binary.ByteOrder, data []byte) error {
	buf := bytes.NewBuffer(data)
	// TODO: use an errReader instead of all these error checks and returns

	var id byte

	if err := binary.Read(buf, order, &id); err != nil {
		return err
	}

	if id != byte(cmd.ID()) {
		return fmt.Errorf("expected ID 0x%X, but found 0x%X", cmd.ID(), id)
	}

	var seed uint32
	err := binary.Read(buf, order, &seed)

	if err != nil {
		return err
	}

	var major, minor, patch, revision uint32

	err = binary.Read(buf, order, &major)
	err = binary.Read(buf, order, &minor)
	err = binary.Read(buf, order, &patch)
	err = binary.Read(buf, order, &revision)

	if err != nil {
		return err
	}

	cmd.Seed = avatar.Seed(seed)
	cmd.Version = avatar.Version{major, minor, patch, revision}

	return nil
}
