package binary

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type LoginRequest struct {
	AccountName  string
	Password     string
	NextLoginKey int
}

var _ ReceivableCommand = &LoginRequest{}

func (cmd LoginRequest) ID() CommandID {
	return 0x80
}

func (cmd LoginRequest) Name() string {
	return "Login Request"
}

func (cmd LoginRequest) ExpectedLength() int {
	return 62
}

func (cmd LoginRequest) IsVariableLength() bool {
	return false
}

func (cmd LoginRequest) IsEncrypted() bool {
	return true
}

func (cmd *LoginRequest) UnmarshalBinary(order binary.ByteOrder, data []byte) error {
	buf := bytes.NewBuffer(data)
	// TODO: use errReader instead of err checks and returns

	var id byte

	if err := binary.Read(buf, order, &id); err != nil {
		return err
	}

	if id != byte(cmd.ID()) {
		return fmt.Errorf("expected ID 0x%X, but found 0x%X", cmd.ID(), id)
	}

	name := make([]byte, 30)
	if err := binary.Read(buf, order, &name); err != nil {
		return err
	}

	cmd.AccountName = string(bytes.Trim(name, "\000"))

	password := make([]byte, 30)
	if err := binary.Read(buf, order, &password); err != nil {
		return err
	}

	cmd.Password = string(bytes.Trim(password, "\000"))

	var key byte
	if err := binary.Read(buf, order, &key); err != nil {
		return err
	}

	cmd.NextLoginKey = int(key)

	return nil
}
