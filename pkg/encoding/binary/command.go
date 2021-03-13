package binary

// variableLength indicates that a binary Command's length is not known until
// all data has been added to the Command.
const variableLength = -1

// CommandID uniquely identifies a binary Command.
type CommandID byte

// Command is a avatar.Command implemented as a binary form.
type Command interface {
	ID() CommandID
	Name() string
	ExpectedLength() int
	IsVariableLength() bool
	IsEncrypted() bool
}

type SendableCommand interface {
	Command
	Marshaler
}

type ReceivableCommand interface {
	Command
	Unmarshaler
}

type SendableReceivableCommand interface {
	SendableCommand
	ReceivableCommand
}

// SystemInfoFlag indicates which data, if any, the client should send in
// response to the GameServerList command.
type SystemInfoFlag byte

// System info flags.
const (
	SystemInfoFlagSendVideoCardInfo      SystemInfoFlag = 0x64
	SystemInfoFlagDoNotSendVideoCardInfo SystemInfoFlag = 0xCC
	SystemInfoFlagUnknown                SystemInfoFlag = 0x5D
	SystemInfoFlagAll                    SystemInfoFlag = 0xFF
)

// LoginDeniedReason captures the reason the login server has denied a client's
// attempt to log in.
type LoginDeniedReason byte

const (
	LoginDeniedReasonUnableToAuthenticate LoginDeniedReason = iota
	LoginDeniedReasonAccountInUse
	LoginDeniedReasonAccountBlocked
	LoginDeniedReasonInvalidCredentials
	LoginDeniedReasonCommunicationProblem
	LoginDeniedReasonConcurrencyLimitMet
	LoginDeniedReasonTimeLimitMet
	LoginDeniedReasonGeneralAuthenticationFailure
	LoginDeniedReasonCouldNotAttachToGameServer
	LoginDeniedReasonCharacterTransferInProgress
)
