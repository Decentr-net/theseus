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

	"github.com/Decentr-net/theseus/internal/entities"
	"github.com/Decentr-net/theseus/internal/service"
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
	//   enum: created_at,likes
	//   example: likes
	// - name: order_by
	//   description: sets sort's direct
	//   in: query
	//   required: false
	//   default: desc
	//   enum: asc,desc
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

	p, err := s.s.GetPost(r.Context(), service.PostID{Owner: owner, UUID: uuid})
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
func extractListParamsFromQuery(q url.Values) (service.ListPostsParams, error) {
	out := service.ListPostsParams{
		SortBy:  service.CreatedAtSortType,
		OrderBy: service.DescendingOrder,
		Limit:   defaultLimit,
	}

	switch s := q.Get("sort_by"); s {
	case createdAtSort:
		out.SortBy = service.CreatedAtSortType
	case likesSort:
		out.SortBy = service.LikesSortType
	case "":
	default:
		return service.ListPostsParams{}, fmt.Errorf("%w: invalid sort_by", errInvalidRequest)
	}

	switch s := q.Get("order_by"); s {
	case ascendingOrder:
		out.OrderBy = service.AscendingOrder
	case descendingOrder:
		out.OrderBy = service.DescendingOrder
	case "":
	default:
		return service.ListPostsParams{}, fmt.Errorf("%w: invalid order_by", errInvalidRequest)
	}

	if s := q.Get("category"); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return service.ListPostsParams{}, fmt.Errorf("%w: failed to parse category", errInvalidRequest)
		}

		c := community.Category(v)
		if c == community.UndefinedCategory || c > community.SportsCategory {
			return service.ListPostsParams{}, fmt.Errorf("%w: invalid category value", errInvalidRequest)
		}
		out.Category = &c
	}

	if s := q.Get("limit"); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return service.ListPostsParams{}, fmt.Errorf("%w: failed to parse limit", errInvalidRequest)
		}

		if v > maxLimit {
			return service.ListPostsParams{}, fmt.Errorf("%w: limit is too big", errInvalidRequest)
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
			return service.ListPostsParams{}, fmt.Errorf("%w: invalid post id", errInvalidRequest)
		}

		out.After = &service.PostID{
			Owner: p[0],
			UUID:  p[1],
		}
	}

	if s := q.Get("from"); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return service.ListPostsParams{}, fmt.Errorf("%w: failed to parse from", errInvalidRequest)
		}

		out.From = &v
	}

	if s := q.Get("to"); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return service.ListPostsParams{}, fmt.Errorf("%w: failed to parse from", errInvalidRequest)
		}

		out.To = &v
	}

	return out, nil
}

func extractProfileIDsFromPosts(p []entities.CalculatedPost) []string {
	out := make([]string, 0, len(p))
	m := make(map[string]struct{}, len(p))

	for i := range p {
		if _, ok := m[p[i].Owner]; !ok {
			out = append(out, p[i].Owner)
			m[p[i].Owner] = struct{}{}
		}
	}

	return out
}

func extractPostIDsFromPosts(p []entities.CalculatedPost) []service.PostID {
	out := make([]service.PostID, len(p))

	for i := range p {
		out[i] = service.PostID{
			Owner: p[i].Owner,
			UUID:  p[i].UUID,
		}
	}

	return out
}

func newListPostsResponse(
	posts []entities.CalculatedPost,
	profiles []entities.Profile,
	stats map[service.PostID]entities.Stats,
) ListPostsResponse {
	out := ListPostsResponse{}

	out.Posts = make([]Post, len(posts))
	for i := range posts {
		out.Posts[i] = Post{
			UUID:         posts[i].UUID,
			Owner:        posts[i].Owner,
			Title:        posts[i].Title,
			Category:     posts[i].Category,
			PreviewImage: posts[i].PreviewImage,
			Text:         posts[i].Text,
			Likes:        posts[i].Likes,
			Dislikes:     posts[i].Dislikes,
			PDV:          posts[i].PDV,
			CreatedAt:    uint64(posts[i].CreatedAt.Unix()),
		}
	}

	out.Profiles = make(map[string]Profile, len(out.Profiles))
	for i := range profiles {
		out.Profiles[profiles[i].Address] = Profile{
			Address:   profiles[i].Address,
			FirstName: profiles[i].FirstName,
			LastName:  profiles[i].LastName,
			Avatar:    profiles[i].Avatar,
			Gender:    profiles[i].Gender,
			Birthday:  profiles[i].Birthday,
			CreatedAt: uint64(profiles[i].CreatedAt.Unix()),
		}
	}

	out.Stats = make(map[string]Stats, len(stats))

	for k, v := range stats {
		s := make(Stats, len(v))

		for k, v := range v {
			s[k.Format(rfc3339date)] = v
		}

		out.Stats[fmt.Sprintf("%s/%s", k.Owner, k.UUID)] = s
	}

	return out
}
