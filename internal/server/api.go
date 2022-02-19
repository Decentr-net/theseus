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
	// ProfileStats contains profiles stats.
	ProfileStats map[string]ProfileStats `json:"profileStats"`
	// Posts' statistics dictionary where key is a full form ID (owner/uuid) and value is statistics
	Stats map[string][]StatsItem `json:"stats"`
}

// GetPostResponse ...
// swagger:model
type GetPostResponse struct {
	Post         Post         `json:"post"`
	ProfileStats ProfileStats `json:"profileStats"`
	Stats        []StatsItem  `json:"stats"`
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
	Slug          string                `json:"slug"`
	LikeWeight    *community.LikeWeight `json:"likeWeight,omitempty"`
	CreatedAt     uint64                `json:"createdAt"`
}

// SharePost ...
// swagger:model
type SharePost struct {
	UUID  string `json:"uuid"`
	Owner string `json:"owner"`
	Title string `json:"title"`
}

// ProfileStats ...
// swagger:model
type ProfileStats struct {
	PostsCount uint16      `json:"postsCount"`
	Stats      []StatsItem `json:"stats"`
}

// DDVStats ...
// swagger:model
type DDVStats struct {
	Total int64       `json:"total"`
	Stats []StatsItem `json:"stats"`
}

// DecentrStats ...
// swagger:model
type DecentrStats struct {
	ADV float64 `json:"adv"`
	DDV float64 `json:"ddv"`
}

// StatsItem ...
// Key is RFC3999 date, value is PDV.
type StatsItem struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}
