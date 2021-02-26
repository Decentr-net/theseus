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
	Posts []Post `json:"posts"`
	// Profiles dictionary where key is an address and value is a profile.
	Profiles map[string]Profile `json:"profiles"`
	// Posts' statistics dictionary where key is a full form ID (owner/uuid) and value is statistics
	Stats map[string]Stats `json:"stats"`
}

// Post ...
type Post struct {
	UUID         string             `json:"uuid"`
	Owner        string             `json:"owner"`
	Title        string             `json:"title"`
	Category     community.Category `json:"category"`
	PreviewImage string             `json:"preview_image"`
	Text         string             `json:"text"`
	Likes        uint32             `json:"likes"`
	Dislikes     uint32             `json:"dislikes"`
	PDV          int64              `json:"pdv"`
	CreatedAt    uint64             `json:"created_at"`
}

// Profile ...
type Profile struct {
	Address   string `json:"address"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
	Avatar    string `json:"avatar"`
	Gender    string `json:"gender"`
	Birthday  string `json:"birthday"`
	CreatedAt uint64 `json:"created_at"`
}

// Stats ...
// Key is RFC3999 date, value is uPDV.
type Stats map[string]int64
