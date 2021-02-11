package storage

import (
	"context"
	"fmt"
	"time"

	community "github.com/Decentr-net/decentr/x/community/types"

	"github.com/Decentr-net/theseus/internal/entities"
)

//go:generate mockgen -destination=./storage_mock.go -package=storage -source=storage.go

// ErrNotFound ...
var ErrNotFound = fmt.Errorf("not found")

// Storage provides methods for interacting with database.
type Storage interface {
	CreatePost(ctx context.Context, p *entities.Post) error
	DeletePost(ctx context.Context, owner string, id string, timestamp time.Time, deletedBy string) error
	SetLike(ctx context.Context, owner string, id string, weight community.LikeWeight, timestamp time.Time, likeOwner string) error
}
