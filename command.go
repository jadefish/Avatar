package avatar

// Command ID bytes.
const (
	CommandLoginDenied    byte = 0x82
	CommandLoginRequest   byte = 0x80
	CommandGameServerList byte = 0xA8
	CommandLoginSeed      byte = 0xEF
)

// A Command is a formatted representation of a Packet.
type Command interface {
	marshallable

	ID() byte
	Name() string
	Length() int

	Execute(Client, Server) error
}

// BaseCommand implements common logic shared among all Commands.
type BaseCommand struct {
	desc Descriptor
	data []byte
}

// NewBaseCommand creates a new base command from the provided descriptor
// and data.
func NewBaseCommand(desc Descriptor, data []byte) *BaseCommand {
	return &BaseCommand{
		desc: desc,
		data: data,
	}
}

// ID returns the Command's unique byte indentifier.
func (cmd BaseCommand) ID() byte {
	return cmd.desc.ID()
}

// Name returns the Command's human-readable string identifier.
func (cmd BaseCommand) Name() string {
	return cmd.desc.Name()
}

// Length returns the length of the Command's data payload.
func (cmd BaseCommand) Length() int {
	return len(cmd.data)
}

// MarshalBinary encodes the Command into a binary form and returns the result.
func (cmd BaseCommand) MarshalBinary() ([]byte, error) {
	panic("command cannot be marshalled")
}

// UnmarshalBinary decodes a binary representation of a Command.
func (cmd BaseCommand) UnmarshalBinary(data []byte) error {
	panic("command cannot be unmarshalled")
}
