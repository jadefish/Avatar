package avatar

// Server is capable of accepting and processing clients.
type Server interface {
	Start() error
	Stop() error

	AccountService() AccountService
	PasswordService() PasswordService
	ShardService() ShardService
}
