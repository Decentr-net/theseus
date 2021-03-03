package server

import (
	community "github.com/Decentr-net/decentr/x/community/types"
)

const maxLimit = 100
const defaultLimit = 20

// Error ...
// swagger:model
type Error struct {
	Error string `json:"error"`
}

// ListPostsResponse ...
// swagger:model
type ListPostsResponse struct {
	Posts []*Post `json:"posts"`
	// Profiles dictionary where key is an address and value is a profile.
	Profiles map[string]Profile `json:"profiles"`
	// Posts' statistics dictionary where key is a full form ID (owner/uuid) and value is statistics
	Stats map[string]Stats `json:"stats"`
}

// GetPostResponse ...
type GetPostResponse struct {
	Post    Post     `json:"post"`
	Profile *Profile `json:"profile"`
	Stats   Stats    `json:"stats"`
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
}

// Stats ...
// Key is RFC3999 date, value is PDV.
type Stats map[string]float64
