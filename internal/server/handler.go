package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi"

	community "github.com/Decentr-net/decentr/x/community/types"

	"github.com/Decentr-net/theseus/internal/storage"
)

var errInvalidRequest = errors.New("invalid request")

func (s server) listPosts(w http.ResponseWriter, r *http.Request) {
	// swagger:operation GET /posts Community ListPosts
	//
	// Return posts with additional meta information.
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: category
	//   description: filters posts by category
	//   in: query
	//   required: false
	//   minimum: 1
	//   maximum: 9
	//   example: 4
	// - name: sort_by
	//   description: sets posts' field to be sorted by
	//   in: query
	//   required: false
	//   default: created_at
	//   type: string
	//   enum: [created_at, likes, dislikes, pdv]
	//   example: likes
	// - name: order_by
	//   description: sets sort's direct
	//   in: query
	//   required: false
	//   default: desc
	//   type: string
	//   enum: [asc, desc]
	//   example: asc
	// - name: owner
	//   description: filters posts by owner
	//   in: query
	//   required: false
	//   example: decentr1ltx6yymrs8eq4nmnhzfzxj6tspjuymh8mgd6gz
	// - name: liked_by
	//   descriptions: filters posts by one who liked its
	//   in: query
	//   required: false
	//   example: decentr1ltx6yymrs8eq4nmnhzfzxj6tspjuymh8mgd6gz
	// - name: limit
	//   description: limits count of returned posts
	//   in: query
	//   required: false
	//   default: 20
	//   minimum: 1
	//   maximum: 100
	// - name: after
	//   description: sets not-including bound for list by post id(`owner/uuid`)
	//   in: query
	//   required: false
	//   example: decentr1ltx6yymrs8eq4nmnhzfzxj6tspjuymh8mgd6gz/df870e39-6fcb-11eb-9461-0242ac11000b
	// - name: from
	//   description: sets lower datetime bound for list
	//   in: query
	//   required: false
	//   example: 1613414389
	// - name: to
	//   description: sets upper datetime bound for list
	//   in: query
	//   required: false
	//   example: 1613424389
	// responses:
	//   '200':
	//     description: Posts
	//     schema:
	//       "$ref": "#/definitions/ListPostsResponse"
	//   '400':
	//     description: bad request
	//     schema:
	//       "$ref": "#/definitions/Error"
	//   '500':
	//     description: internal server error
	//     schema:
	//       "$ref": "#/definitions/Error"

	params, err := extractListParamsFromQuery(r.URL.Query())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	posts, err := s.s.ListPosts(r.Context(), params)
	if err != nil {
		writeInternalError(getLogger(r.Context()), w, err.Error())
		return
	}

	profiles, err := s.s.GetProfiles(r.Context(), extractProfileIDsFromPosts(posts))
	if err != nil {
		writeInternalError(getLogger(r.Context()), w, err.Error())
		return
	}

	stats, err := s.s.GetStats(r.Context(), extractPostIDsFromPosts(posts))
	if err != nil {
		writeInternalError(getLogger(r.Context()), w, err.Error())
		return
	}

	writeOK(w, http.StatusOK, newListPostsResponse(posts, profiles, stats))
}

func (s server) getPost(w http.ResponseWriter, r *http.Request) {
	// swagger:operation GET /posts/{owner}/{uuid} Community GetPost
	//
	// Get post by owner and uuid.
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   required: true
	//   type: string
	// - name: uuid
	//   in: path
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: Posts
	//     schema:
	//       "$ref": "#/definitions/ListPostsResponse"
	//   '400':
	//     description: bad request
	//     schema:
	//       "$ref": "#/definitions/Error"
	//   '500':
	//     description: internal server error
	//     schema:
	//       "$ref": "#/definitions/Error"

	owner, uuid := chi.URLParam(r, "owner"), chi.URLParam(r, "uuid")

	if owner == "" || uuid == "" {
		writeError(w, http.StatusBadRequest, "invalid owner or uuid")
		return
	}

	p, err := s.s.GetPost(r.Context(), storage.PostID{Owner: owner, UUID: uuid})
	if err != nil {
		writeInternalError(getLogger(r.Context()).WithError(err), w, "failed to get post")
		return
	}

	writeOK(w, http.StatusOK, Post{
		UUID:         p.Owner,
		Owner:        p.Owner,
		Title:        p.Title,
		Category:     p.Category,
		PreviewImage: p.PreviewImage,
		Text:         p.Text,
		Likes:        p.Likes,
		Dislikes:     p.Dislikes,
		PDV:          p.PDV,
		CreatedAt:    uint64(p.CreatedAt.Unix()),
	})
}

// nolint: gocyclo
func extractListParamsFromQuery(q url.Values) (*storage.ListPostsParams, error) {
	out := storage.ListPostsParams{
		SortBy:  storage.CreatedAtSortType,
		OrderBy: storage.DescendingOrder,
		Limit:   defaultLimit,
	}

	sortBy := storage.SortType(strings.ToLower(q.Get("sort_by")))
	switch sortBy {
	case storage.CreatedAtSortType, storage.LikesSortType, storage.DislikesSortType, storage.PDVSortType:
		out.SortBy = sortBy
	case "":
	default:
		return nil, fmt.Errorf("%w: invalid sort_by", errInvalidRequest)
	}

	orderBy := storage.OrderType(strings.ToLower(q.Get("order_by")))
	switch orderBy {
	case storage.AscendingOrder, storage.DescendingOrder:
		out.OrderBy = orderBy
	case "":
	default:
		return nil, fmt.Errorf("%w: invalid order_by", errInvalidRequest)
	}

	if s := q.Get("category"); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to parse category", errInvalidRequest)
		}

		c := community.Category(v)
		if c == community.UndefinedCategory || c > community.SportsCategory {
			return nil, fmt.Errorf("%w: invalid category value", errInvalidRequest)
		}
		out.Category = &c
	}

	if s := q.Get("limit"); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to parse limit", errInvalidRequest)
		}

		if v > maxLimit {
			return nil, fmt.Errorf("%w: limit is too big", errInvalidRequest)
		}

		out.Limit = uint16(v)
	}

	if s := q.Get("owner"); s != "" {
		out.Owner = &s
	}

	if s := q.Get("liked_by"); s != "" {
		out.LikedBy = &s
	}

	if s := q.Get("after"); s != "" {
		p := strings.Split(s, "/")

		if len(p) != 2 {
			return nil, fmt.Errorf("%w: invalid post id", errInvalidRequest)
		}

		out.After = &storage.PostID{
			Owner: p[0],
			UUID:  p[1],
		}
	}

	if s := q.Get("from"); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to parse from", errInvalidRequest)
		}

		out.From = &v
	}

	if s := q.Get("to"); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to parse from", errInvalidRequest)
		}

		out.To = &v
	}

	return &out, nil
}

func extractProfileIDsFromPosts(p []*storage.Post) []string {
	out := make([]string, 0, len(p))
	m := make(map[string]struct{}, len(p))

	for _, v := range p {
		if _, ok := m[v.Owner]; !ok {
			out = append(out, v.Owner)
			m[v.Owner] = struct{}{}
		}
	}

	return out
}

func extractPostIDsFromPosts(p []*storage.Post) []storage.PostID {
	out := make([]storage.PostID, len(p))

	for i := range p {
		out[i] = storage.PostID{
			Owner: p[i].Owner,
			UUID:  p[i].UUID,
		}
	}

	return out
}

func newListPostsResponse(
	posts []*storage.Post,
	profiles []*storage.Profile,
	stats map[storage.PostID]storage.Stats,
) ListPostsResponse {
	out := ListPostsResponse{}

	out.Posts = make([]Post, len(posts))
	for i, v := range posts {
		out.Posts[i] = Post{
			UUID:         v.UUID,
			Owner:        v.Owner,
			Title:        v.Title,
			Category:     v.Category,
			PreviewImage: v.PreviewImage,
			Text:         v.Text,
			Likes:        v.Likes,
			Dislikes:     v.Dislikes,
			PDV:          v.PDV,
			CreatedAt:    uint64(v.CreatedAt.Unix()),
		}
	}

	out.Profiles = make(map[string]Profile, len(out.Profiles))
	for _, v := range profiles {
		out.Profiles[v.Address] = Profile{
			Address:   v.Address,
			FirstName: v.FirstName,
			LastName:  v.LastName,
			Avatar:    v.Avatar,
			Gender:    v.Gender,
			Birthday:  v.Birthday,
			CreatedAt: uint64(v.CreatedAt.Unix()),
		}
	}

	out.Stats = make(map[string]Stats, len(stats))

	for k, v := range stats {
		out.Stats[fmt.Sprintf("%s/%s", k.Owner, k.UUID)] = Stats(v)
	}

	return out
}
