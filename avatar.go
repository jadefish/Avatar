package avatar

import (
	"fmt"
	"strconv"
)

type Version struct {
	Major    uint32
	Minor    uint32
	Patch    uint32
	Revision uint32
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Patch, v.Revision)
}

// EntityID represents a unique identifier for a domain entity.
type EntityID int64

func (id EntityID) String() string {
	return strconv.FormatInt(int64(id), 10)
}

// Valid determines whether the ID is valid.
func (id EntityID) Valid() bool {
	return id > 0
}
