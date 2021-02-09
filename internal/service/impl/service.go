// Package impl is implementation of service interface.
package impl

import (
	"context"
	"time"

	"github.com/Decentr-net/decentr/x/community/types"

	"github.com/Decentr-net/theseus/internal/entities"
	"github.com/Decentr-net/theseus/internal/service"
	"github.com/Decentr-net/theseus/internal/storage"
)

// service ...
type srv struct {
	storage storage.Storage
}

func (s srv) OnHeight(height uint64, f func(s service.Service) error) error {
	panic("implement me")
}

func (s srv) GetHeight() (uint64, error) {
	panic("implement me")
}

func (s srv) CreatePost(ctx context.Context, p *entities.Post) error {
	panic("implement me")
}

func (s srv) DeletePost(ctx context.Context, postOwner, postUUID string, timestamp time.Time, deletedByUUID string) error {
	panic("implement me")
}

func (s srv) SetLike(ctx context.Context, postOnwer, postUUID string, weight types.LikeWeight, timestamp time.Time, likedBy string) error {
	panic("implement me")
}

// New creates new instance of service.
func New(storage storage.Storage) service.Service {
	return srv{
		storage: storage,
	}
}
