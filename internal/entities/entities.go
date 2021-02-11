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
