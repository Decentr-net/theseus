// Package storage contains a storage interface.
package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	community "github.com/Decentr-net/decentr/x/community/types"

	"github.com/Decentr-net/theseus/internal/entities"
)

//go:generate mockgen -destination=./mock/storage.go -package=mock -source=storage.go

// ErrNotFound ...
var ErrNotFound = fmt.Errorf("not found")

// ErrRequestedHeightIsTooHigh returned when the height requested in WithLockedHeight function is more than expected.
var ErrRequestedHeightIsTooHigh = errors.New("requested height is too high")

// ErrRequestedHeightIsTooLow returned when the height requested in WithLockedHeight function is less than expected.
var ErrRequestedHeightIsTooLow = errors.New("requested height is too low")

// Storage provides methods for interacting with database.
type Storage interface {
	WithLockedHeight(ctx context.Context, height uint64, f func(s Storage) error) error
	GetHeight(ctx context.Context) (uint64, error)
	CreatePost(ctx context.Context, p *entities.Post) error
	GetPost(ctx context.Context, owner, uuid string) (*entities.Post, error)
	DeletePost(ctx context.Context, postOwner string, postUUID string, timestamp time.Time, deletedBy string) error
	SetLike(ctx context.Context, postOwner string, postUUID string, weight community.LikeWeight, timestamp time.Time, likeOwner string) error
}
