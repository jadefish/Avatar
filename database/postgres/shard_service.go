package postgres

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/jadefish/avatar"
)

var errInvalidShardName = errors.New("invalid shard name")

// ShardService facilitates interacting with shards.
type ShardService struct {
	DB *sqlx.DB
}

// All retrieves all non-deleted shards.
func (s ShardService) All() ([]avatar.Shard, error) {
	shards := make([]avatar.Shard, 0, 10)

	err := s.DB.Get(shards, `
		SELECT s.*
		FROM shards s
		WHERE s.deleted_at IS NULL
		ORDER BY s.name ASC;
	`)

	if err != nil {
		return nil, errors.Wrap(err, "all")
	}

	return shards, nil
}

// Find a non-deleted shard by name.
func (s ShardService) Find(name string) (*avatar.Shard, error) {
	if len(name) < 1 || len(name) > avatar.ShardNameLength {
		return nil, errors.Wrap(errInvalidShardName, "find")
	}

	shard := &avatar.Shard{}

	err := s.DB.Get(shard, `
		SELECT s.*
		FROM shards
		WHERE s.name = ?
		AND s.deleted_at IS NULL
		ORDER BY s.created_at DESC
		LIMIT 1;
	`, name)

	if err != nil {
		return nil, errors.Wrap(err, "find")
	}

	return shard, nil
}
