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

// SortType ...
type SortType uint8

const (
	// CreatedAtSortType ...
	CreatedAtSortType SortType = iota
	// LikesSortType ...
	LikesSortType
)

// OrderType ...
type OrderType uint8

const (
	// AscendingOrder ...
	AscendingOrder OrderType = iota
	// DescendingOrder ...
	DescendingOrder
)

// Service ...
type Service interface {
	OnHeight(ctx context.Context, height uint64, f func(s Service) error) error
	GetHeight(ctx context.Context) (uint64, error)

	GetProfiles(ctx context.Context, addr []string) ([]entities.Profile, error)
	SetProfile(ctx context.Context, p *entities.Profile) error

	Follow(ctx context.Context, follower, followee string) error
	Unfollow(ctx context.Context, follower, followee string) error

	ListPosts(ctx context.Context, filter ListPostsParams) ([]entities.CalculatedPost, error)
	GetPost(ctx context.Context, id PostID) (entities.CalculatedPost, error)
	CreatePost(ctx context.Context, p *entities.Post) error
	DeletePost(ctx context.Context, postOwner, postUUID string, timestamp time.Time, deletedByUUID string) error
	SetLike(ctx context.Context, postOwner, postUUID string, weight community.LikeWeight, timestamp time.Time, likedBy string) error

	GetStats(ctx context.Context, id []PostID) (map[PostID]entities.Stats, error)
}

// ListPostsParams ...
type ListPostsParams struct {
	SortBy   SortType
	OrderBy  OrderType
	Limit    uint16
	Category *community.Category
	Owner    *string
	LikedBy  *string
	After    *PostID
	From     *uint64
	To       *uint64
}

// PostID ...
type PostID struct {
	Owner string
	UUID  string
}
