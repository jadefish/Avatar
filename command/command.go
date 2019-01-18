package command

import (
	"bytes"
	"errors"
	"io"
	"log"
	"reflect"

	"github.com/jadefish/avatar"
)

var factory = map[byte]avatar.Command{
	0x80: (*LoginRequest)(nil),
	0xEF: (*LoginSeed)(nil),
}

var errNoBuilder = errors.New("command cannot be built: no builder found")

// Make constructs a new Command from a descriptor, reading and
// unmarshalling all available data from the Reader r as the source of data
// for instantiation.
func Make(desc avatar.Descriptor, r io.Reader) (avatar.Command, error) {
	prototype, ok := factory[desc.ID()]

	if !ok {
		return nil, errNoBuilder
	}

	var b bytes.Buffer
	_, err := io.Copy(&b, r)

	if err != nil {
		return nil, err
	}

	data := b.Bytes()
	base := avatar.NewBaseCommand(desc, data)

	// Construct the specific command:
	ptr := reflect.New(reflect.TypeOf(prototype).Elem())
	elem := ptr.Elem()
	elem.FieldByName("BaseCommand").Set(reflect.ValueOf(*base))
	cmd := ptr.Interface().(avatar.Command)
	cmd.UnmarshalBinary(data)

	log.Printf("made command of type %T: %+v\n", cmd, cmd)

	return cmd, nil
}
