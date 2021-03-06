// Package storage contains a storage interface.
package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	community "github.com/Decentr-net/decentr/x/community/types"
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

	GetProfiles(ctx context.Context, addr ...string) ([]*Profile, error)
	SetProfile(ctx context.Context, p *SetProfileParams) error

	Follow(ctx context.Context, follower, followee string) error
	Unfollow(ctx context.Context, follower, followee string) error

	ListPosts(ctx context.Context, p *ListPostsParams) ([]*Post, error)
	CreatePost(ctx context.Context, p *CreatePostParams) error
	GetPost(ctx context.Context, id PostID) (*Post, error)
	DeletePost(ctx context.Context, id PostID, timestamp time.Time, deletedBy string) error

	GetLikes(ctx context.Context, likedBy string, id ...PostID) (map[PostID]community.LikeWeight, error)
	SetLike(ctx context.Context, id PostID, weight community.LikeWeight, timestamp time.Time, likeOwner string) error

	GetStats(ctx context.Context, id ...PostID) (map[PostID]Stats, error)

	AddPDV(ctx context.Context, address string, amount int64, timestamp time.Time) error
	GetProfileStats(ctx context.Context, address string) (Stats, error)
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

// SetProfileParams ...
type SetProfileParams struct {
	Address   string
	FirstName string
	LastName  string
	Bio       string
	Avatar    string
	Gender    string
	Birthday  string
	CreatedAt time.Time
}

// Profile ...
type Profile struct {
	Address    string
	FirstName  string
	LastName   string
	Bio        string
	Avatar     string
	Gender     string
	Birthday   string
	CreatedAt  time.Time
	PostsCount uint16
}

// Stats is map where key is date in RFC3339 format and value is uPDV count.
type Stats map[string]int64
