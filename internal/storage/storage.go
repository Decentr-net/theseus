// Package storage contains a storage interface.
package storage

import (
	"context"
	"fmt"
	"time"

	community "github.com/Decentr-net/decentr/x/community/types"
)

//go:generate mockgen -destination=./mock/storage.go -package=mock -source=storage.go

// ErrNotFound ...
var ErrNotFound = fmt.Errorf("not found")

// Storage provides methods for interacting with database.
type Storage interface {
	InTx(ctx context.Context, f func(s Storage) error) error
	SetHeight(ctx context.Context, height uint64) error
	GetHeight(ctx context.Context) (uint64, error)
	RefreshViews(ctx context.Context) error

	Follow(ctx context.Context, follower, followee string) error
	Unfollow(ctx context.Context, follower, followee string) error

	ListPosts(ctx context.Context, p *ListPostsParams) ([]*Post, error)
	CreatePost(ctx context.Context, p *CreatePostParams) error
	GetPost(ctx context.Context, id PostID) (*Post, error)
	DeletePost(ctx context.Context, id PostID, timestamp time.Time, deletedBy string) error

	GetLikes(ctx context.Context, likedBy string, id ...PostID) (map[PostID]community.LikeWeight, error)
	SetLike(ctx context.Context, id PostID, weight community.LikeWeight, timestamp time.Time, likeOwner string) error

	AddPDV(ctx context.Context, address string, amount int64, timestamp time.Time) error

	GetProfileStats(ctx context.Context, addr ...string) ([]*ProfileStats, error)
	GetPostStats(ctx context.Context, id ...PostID) (map[PostID]Stats, error)

	GetDecentrStats(ctx context.Context) (*DecentrStats, error)
}

// SortType ...
type SortType string

const (
	// CreatedAtSortType ...
	CreatedAtSortType SortType = "created_at"
	// LikesSortType ...
	LikesSortType SortType = "likes"
	// DislikesSortType ...
	DislikesSortType SortType = "dislikes"
	// PDVSortType ...
	PDVSortType SortType = "updv"
)

// OrderType ...
type OrderType string

const (
	// AscendingOrder ...
	AscendingOrder OrderType = "asc"
	// DescendingOrder ...
	DescendingOrder = "desc"
)

// ListPostsParams ...
type ListPostsParams struct {
	SortBy     SortType
	OrderBy    OrderType
	Limit      uint16
	Category   *community.Category
	Owner      *string
	LikedBy    *string
	FollowedBy *string
	After      *PostID
	From       *uint64
	To         *uint64
}

// PostID ...
type PostID struct {
	Owner string
	UUID  string
}

// CreatePostParams ...
type CreatePostParams struct {
	UUID         string
	Owner        string
	Title        string
	Category     community.Category
	PreviewImage string
	Text         string
	CreatedAt    time.Time
}

// Post ...
type Post struct {
	UUID         string
	Owner        string
	Title        string
	Category     community.Category
	PreviewImage string
	Text         string
	CreatedAt    time.Time
	Likes        uint32
	Dislikes     uint32
	UPDV         int64
}

// ProfileStats ...
type ProfileStats struct {
	Address    string
	PostsCount uint16
	Stats      Stats
}

// DecentrStats represents all users stats.
type DecentrStats struct {
	ADV float64 // Average earned pdv
	DDV int64   // Whole earned pdv
}

// Stats is map where key is date in RFC3339 format and value is uPDV count.
type Stats map[string]int64
