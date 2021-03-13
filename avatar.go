package avatar

import (
	"strconv"
)

// EntityID represents a unique identifier for a domain entity.
type EntityID int64

func (id EntityID) String() string {
	return strconv.FormatInt(int64(id), 10)
}

// Valid determines whether the ID is valid.
func (id EntityID) Valid() bool {
	return id > 0
}
