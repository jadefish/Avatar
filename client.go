package avatar

import (
	"fmt"
)

// Version identifies a unique client revision.
type Version struct {
	Major    uint32
	Minor    uint32
	Patch    uint32
	Revision uint32
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Patch, v.Revision)
}

// ClientState captures the current state of the client.
type ClientState int

const (
	// ClientStateUnknown indicates that the current state of the client is not
	// currently known. It is not safe to assume the client can be used.
	ClientStateUnknown ClientState = iota

	// ClientStateDisconnected indicates that the client's connection to the
	// server is no longer valid.
	ClientStateDisconnected

	// ClientStateNew indicates that a client has just recently connected and
	// has not yet been authenticated.
	ClientStateNew

	// ClientStateAuthenticated describes a client has been successfully
	// authenticated.
	ClientStateAuthenticated
)
