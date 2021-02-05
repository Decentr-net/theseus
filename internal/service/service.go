package service

import (
	"context"
	"time"

	community "github.com/Decentr-net/decentr/x/community/types"

	"github.com/Decentr-net/theseus/internal/storage"
)

//go:generate mockgen -destination=./service_mock.go -package=service -source=service.go

// Post ...
type Post struct {
	UUID         string
	Owner        string
	Title        string
	Category     community.Category
	PreviewImage string
	Text         string
	CreatedAt    time.Time
	DeletedAt    *time.Time
	DeletedBy    *string
}

// Service ...
type Service interface {
	CreatePost(ctx context.Context, p Post) error
	DeletePost(ctx context.Context, owner string, id string, timestamp time.Time, deletedBy string) error
	SetLike(ctx context.Context, owner string, id string, weight community.LikeWeight, timestamp time.Time, likeOwner string) error
}

// Service ...
type service struct {
	storage storage.Storage
}

// New creates new instance of service.
func New(storage storage.Storage) Service {
	return service{
		storage: storage,
	}
}

func (s service) CreatePost(ctx context.Context, p Post) error {
	panic("implement me")
}

func (s service) DeletePost(ctx context.Context, owner string, id string, timestamp time.Time, deletedBy string) error {
	panic("implement me")
}

func (s service) SetLike(ctx context.Context, owner string, id string, weight community.LikeWeight, timestamp time.Time, likeOwner string) error {
	panic("implement me")
}
