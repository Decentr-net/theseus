package server

import (
	community "github.com/Decentr-net/decentr/x/community/types"
)

const maxLimit = 100
const defaultLimit = 20

// ListPostsResponse ...
// swagger:model
type ListPostsResponse struct {
	Posts []*Post `json:"posts"`
	// Profiles dictionary where key is an address and value is a profile.
	Profiles map[string]Profile `json:"profiles"`
	// Posts' statistics dictionary where key is a full form ID (owner/uuid) and value is statistics
	Stats map[string][]StatsItem `json:"stats"`
}

// GetPostResponse ...
// swagger:model
type GetPostResponse struct {
	Post    Post        `json:"post"`
	Profile *Profile    `json:"profile"`
	Stats   []StatsItem `json:"stats"`
}

// Post ...
type Post struct {
	UUID          string                `json:"uuid"`
	Owner         string                `json:"owner"`
	Title         string                `json:"title"`
	Category      community.Category    `json:"category"`
	PreviewImage  string                `json:"previewImage"`
	Text          string                `json:"text"`
	LikesCount    uint32                `json:"likesCount"`
	DislikesCount uint32                `json:"dislikesCount"`
	PDV           float64               `json:"pdv"`
	LikeWeight    *community.LikeWeight `json:"likeWeight,omitempty"`
	CreatedAt     uint64                `json:"createdAt"`
}

// Profile ...
type Profile struct {
	Address      string `json:"address"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Bio          string `json:"bio"`
	Avatar       string `json:"avatar"`
	Gender       string `json:"gender"`
	Birthday     string `json:"birthday"`
	RegisteredAt uint64 `json:"registeredAt"`

	PostsCount uint16 `json:"postsCount"`
}

// AllStats ...
// swagger:model
type AllStats struct {
	ADV float64 `json:"adv"`
	DDV int64   `json:"ddv"`
}

// StatsItem ...
// Key is RFC3999 date, value is PDV.
type StatsItem struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}
