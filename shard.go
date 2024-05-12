package avatar

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

type ShardIdentifier int

type Shard struct {
	ID       ShardIdentifier
	Name     string
	Index    int
	TimeZone *time.Location
	Capacity int
	Public   bool
}

type ShardLoader interface {
	Get(ShardIdentifier) (Shard, error)
}

type ShardWriter interface {
	Store(Shard) error
}

type ShardRemover interface {
	Remove(ShardIdentifier) error
}

var ErrMissingEntity = errors.New("entity does not exist")

type RegisterShard struct {
	loader ShardLoader
	writer ShardWriter
}
type DeregisterShard struct {
	remover ShardRemover
}

// https://adodd.net/post/go-ddd-repository-pattern

func (s Shard) Register(cmd RegisterShard) error {
	_, err := cmd.loader.Get(s.ID)

	if err == nil {
		return errors.New("shard has already been registered")
	}

	name := strings.TrimSpace(s.Name)
	if utf8.RuneCountInString(name) < 1 {
		return errors.New("shard name cannot be empty")
	}

	if s.Capacity < 0 {
		return errors.New("shard cannot have a negative capacity")
	}

	validatedShard := Shard{
		Name:     strings.TrimSpace(s.Name),
		TimeZone: s.TimeZone,
		Capacity: s.Capacity,
		Public:   s.Public,
	}

	return cmd.writer.Store(validatedShard)
}

func (s Shard) Deregister(cmd DeregisterShard) error {
	return cmd.remover.Remove(s.ID)
}
