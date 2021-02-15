// Package entities contains main entities of service.
package entities

import (
	"time"

	community "github.com/Decentr-net/decentr/x/community/types"
)

// Post ...
type Post struct {
	UUID         string
	Owner        string
	Title        string
	Category     community.Category
	PreviewImage string
	Text         string
	CreatedAt    time.Time
}

// CalculatedPost ...
type CalculatedPost struct {
	Post
	Likes    uint32
	Dislikes uint32
	PDV      uint64
}

// Profile ...
type Profile struct {
	Address   string
	FirstName string
	LastName  string
	Avatar    string
	Gender    string
	Birthday  string
	CreatedAt time.Time
}

// Stats is map where key is date and value is uPDV count.
type Stats map[time.Time]uint64
