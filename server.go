package avatar

// Server is capable of accepting and processing clients.
type Server interface {
	Start() error
	Stop() error
	Process(*Client) error
}
