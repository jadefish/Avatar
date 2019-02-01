package avatar

import (
	"math/rand"
	"net"
	"time"
)

// ShardNameLength is the maximum number of US ASCII characters allowed to
// use as a shard name.
const ShardNameLength = 32

// ShardService provides methods for retrieving shards.
type ShardService interface {
	All() ([]Shard, error)
	Find(name string) (*Shard, error)
}

// Shard contains information about a game server.
type Shard struct {
	Record

	Name      string
	TimeZone  string
	Capacity  int
	IPAddress net.IP
}

// Location retrieves the named geographical location in which the shard
// resides.
func (s Shard) Location() (*time.Location, error) {
	l, err := time.LoadLocation(s.TimeZone)

	if err != nil {
		return nil, err
	}

	return l, nil
}

// PercentFull returns how full the shard is based on the number of clients
// in the game world and the shard's capacity.
func (s Shard) PercentFull() uint {
	// TODO: len(clients) / s.Capacity
	return uint(rand.Uint32())
}
