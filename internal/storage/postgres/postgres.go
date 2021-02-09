// Package postgres is implementation of storage interface.
package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/Decentr-net/decentr/x/community/types"
	"github.com/Decentr-net/theseus/internal/entities"
	"github.com/Decentr-net/theseus/internal/storage"
)

//const uniqueViolationErrorCode = "23505"

type pg struct {
	db *sqlx.DB
}

func (p2 pg) CreatePost(ctx context.Context, p *entities.Post) error {
	panic("implement me")
}

func (p2 pg) DeletePost(ctx context.Context, owner string, id string, timestamp time.Time, deletedBy string) error {
	panic("implement me")
}

func (p2 pg) SetLike(ctx context.Context, owner string, id string, weight types.LikeWeight, timestamp time.Time, likeOwner string) error {
	panic("implement me")
}

// New creates new instance of pg.
func New(db *sql.DB) storage.Storage {
	return pg{
		db: sqlx.NewDb(db, "postgres"),
	}
}
