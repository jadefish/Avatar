package postgres

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/jadefish/avatar"
)

var errInvalidShardName = errors.New("invalid shard name")

// NewShardService creates a new shard service backed by PostgreSQL.
func NewShardService(db *sqlx.DB) *shardService {
	return &shardService{
		db: db,
	}
}

// shardService facilitates interacting with shards.
type shardService struct {
	db *sqlx.DB
}

// All retrieves all non-deleted shards.
func (s shardService) All() ([]avatar.Shard, error) {
	shards := make([]avatar.Shard, 0, 10)

	err := s.db.Get(shards, `
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
func (s shardService) Find(name string) (*avatar.Shard, error) {
	if len(name) < 1 || len(name) > avatar.ShardNameLength {
		return nil, errors.Wrap(errInvalidShardName, "find")
	}

	shard := &avatar.Shard{}

	err := s.db.Get(shard, `
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
