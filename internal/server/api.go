package server

var (
//	errInvalidRequest = errors.New("invalid request")
)

// Error ...
// swagger:model
type Error struct {
	Error string `json:"error"`
}
