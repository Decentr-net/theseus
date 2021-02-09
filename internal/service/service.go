// Package service contains interface for service business-logic.
package service

import (
	"context"
	"time"

	community "github.com/Decentr-net/decentr/x/community/types"

	"github.com/Decentr-net/theseus/internal/entities"
)

//go:generate mockgen -destination=./mock/service.go -package=mock -source=service.go

// Service ...
type Service interface {
	OnHeight(height uint64, f func(s Service) error) error
	GetHeight() (uint64, error)

	CreatePost(ctx context.Context, p *entities.Post) error
	DeletePost(ctx context.Context, postOwner, postUUID string, timestamp time.Time, deletedByUUID string) error
	SetLike(ctx context.Context, postOnwer, postUUID string, weight community.LikeWeight, timestamp time.Time, likedBy string) error
}
