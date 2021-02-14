// Package service contains interface for service business-logic.
package service

import (
	"context"
	"errors"
	"time"

	community "github.com/Decentr-net/decentr/x/community/types"

	"github.com/Decentr-net/theseus/internal/entities"
)

//go:generate mockgen -destination=./mock/service.go -package=mock -source=service.go

// ErrRequestedHeightIsTooLow returned when the height requested in OnHeight function is less than expected.
var ErrRequestedHeightIsTooLow = errors.New("requested height is too low")

// Service ...
type Service interface {
	OnHeight(ctx context.Context, height uint64, f func(s Service) error) error
	GetHeight(ctx context.Context) (uint64, error)

	SetProfile(ctx context.Context, p *entities.Profile) error

	Follow(ctx context.Context, follower, followee string) error
	Unfollow(ctx context.Context, follower, followee string) error

	CreatePost(ctx context.Context, p *entities.Post) error
	DeletePost(ctx context.Context, postOwner, postUUID string, timestamp time.Time, deletedByUUID string) error
	SetLike(ctx context.Context, postOwner, postUUID string, weight community.LikeWeight, timestamp time.Time, likedBy string) error
}
