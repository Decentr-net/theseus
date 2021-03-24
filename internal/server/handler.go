package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/go-chi/chi"

	community "github.com/Decentr-net/decentr/x/community/types"
	token "github.com/Decentr-net/decentr/x/token/types"
	"github.com/Decentr-net/decentr/x/utils"
	"github.com/Decentr-net/go-api"

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
	// - name: sortBy
	//   description: sets posts' field to be sorted by
	//   in: query
	//   required: false
	//   default: createdAt
	//   type: string
	//   enum: [created_at, likesCount, dislikesCount, pdv]
	//   example: likes
	// - name: orderBy
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
	// - name: likedBy
	//   descriptions: filters posts by one who liked its
	//   in: query
	//   required: false
	//   example: decentr1ltx6yymrs8eq4nmnhzfzxj6tspjuymh8mgd6gz
	// - name: followedBy
	//   in: query
	//   description: filters post by owners who followed by followedBy
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
	// - name: requestedBy
	//   in: query
	//   description: adds liked flag to response
	//   required: false
	//   example: decentr1ltx6yymrs8eq4nmnhzfzxj6tspjuymh8mgd6gz
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
		api.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	posts, err := s.s.ListPosts(r.Context(), params)
	if err != nil {
		api.WriteInternalErrorf(r.Context(), w, "failed to list posts: %s", err.Error())
		return
	}

	profiles, err := s.s.GetProfiles(r.Context(), extractProfileIDsFromPosts(posts)...)
	if err != nil {
		api.WriteInternalErrorf(r.Context(), w, "failed to get profiles: %s", err.Error())
		return
	}

	ids := extractPostIDsFromPosts(posts)
	stats, err := s.s.GetStats(r.Context(), ids...)
	if err != nil {
		api.WriteInternalErrorf(r.Context(), w, "failed get stats: %s", err.Error())
		return
	}

	var liked map[storage.PostID]community.LikeWeight
	if requestedBy := r.URL.Query().Get("requestedBy"); requestedBy != "" {
		liked, err = s.s.GetLikes(r.Context(), requestedBy, ids...)
		if err != nil {
			api.WriteInternalErrorf(r.Context(), w, "failed to get likes: %s", err.Error())
			return
		}
	}

	api.WriteOK(w, http.StatusOK, newListPostsResponse(posts, profiles, stats, liked))
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
	// - name: requestedBy
	//   in: query
	//   description: adds liked flag to response
	//   required: false
	//   example: decentr1ltx6yymrs8eq4nmnhzfzxj6tspjuymh8mgd6gz
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
		api.WriteError(w, http.StatusBadRequest, "invalid owner or uuid")
		return
	}

	post, err := s.s.GetPost(r.Context(), storage.PostID{Owner: owner, UUID: uuid})
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			api.WriteError(w, http.StatusNotFound, "post not found")
			return
		}
		api.WriteInternalErrorf(r.Context(), w, "failed to get post: %s", err.Error())
		return
	}

	profile, err := s.s.GetProfiles(r.Context(), post.Owner)
	if err != nil {
		api.WriteInternalErrorf(r.Context(), w, "failed to get profile: %s", err.Error())
		return
	}

	pID := storage.PostID{Owner: post.Owner, UUID: post.UUID}
	stats, err := s.s.GetStats(r.Context(), pID)
	if err != nil {
		api.WriteInternalErrorf(r.Context(), w, "failed to get stats: %s", err.Error())
		return
	}

	resp := GetPostResponse{
		Post: *toAPIPost(post),
	}

	if len(profile) == 1 {
		resp.Profile = toAPIProfile(profile[0])
	}

	if s, ok := stats[pID]; ok {
		resp.Stats = toAPIStats(s)
	}

	if requestedBy := r.URL.Query().Get("requestedBy"); requestedBy != "" {
		liked, err := s.s.GetLikes(r.Context(), requestedBy, pID)
		if err != nil {
			api.WriteInternalErrorf(r.Context(), w, "failed to get like: %s", err.Error())
			return
		}

		v := liked[pID]
		resp.Post.LikeWeight = &v
	}

	api.WriteOK(w, http.StatusOK, resp)
}

func (s server) getProfileStats(w http.ResponseWriter, r *http.Request) {
	// swagger:operation GET /profiles/{address}/stats Profiles GetProfileStats
	//
	// Get pdv stats by address.
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: address
	//   in: path
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: Posts
	//     schema:
	//       type: array
	//       items:
	//         "$ref": "#/definitions/StatsItem"
	//   '404':
	//     description: profile not found
	//     schema:
	//       "$ref": "#/definitions/Error"
	//   '400':
	//     description: bad request
	//     schema:
	//       "$ref": "#/definitions/Error"
	//   '500':
	//     description: internal server error
	//     schema:
	//       "$ref": "#/definitions/Error"

	address := chi.URLParam(r, "address")

	if address == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid address")
		return
	}

	stats, err := s.s.GetProfileStats(r.Context(), address)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			api.WriteOK(w, http.StatusOK, []StatsItem{})
			return
		}
		api.WriteInternalErrorf(r.Context(), w, "failed to get profile stats: %s", err.Error())
		return
	}

	api.WriteOK(w, http.StatusOK, toAPIStats(stats))
}

func (s server) getDecentrStats(w http.ResponseWriter, r *http.Request) {
	// swagger:operation GET /profiles/stats Profiles GetDecentrStats
	//
	// Returns decentr stats.
	//
	// ---
	// produces:
	// - application/json
	// responses:
	//   '200':
	//     description: Stats
	//     schema:
	//       "$ref": "#/definitions/DecentrStats"
	//   '500':
	//     description: internal server error
	//     schema:
	//       "$ref": "#/definitions/Error"

	stats, err := s.s.GetDecentrStats(r.Context())
	if err != nil {
		api.WriteInternalErrorf(r.Context(), w, "failed to get all users stats: %s", err.Error())
		return
	}

	api.WriteOK(w, http.StatusOK, DecentrStats{
		ADV: stats.ADV / float64(token.Denominator),
		DDV: float64(stats.DDV) / float64(token.Denominator),
	})
}

// nolint: gocyclo
func extractListParamsFromQuery(q url.Values) (*storage.ListPostsParams, error) {
	out := storage.ListPostsParams{
		SortBy:  storage.CreatedAtSortType,
		OrderBy: storage.DescendingOrder,
		Limit:   defaultLimit,
	}

	sortBy := q.Get("sortBy")
	switch sortBy {
	case "createdAt":
		out.SortBy = storage.CreatedAtSortType
	case "likesCount":
		out.SortBy = storage.LikesSortType
	case "dislikesCount":
		out.SortBy = storage.DislikesSortType
	case "pdv":
		out.SortBy = storage.PDVSortType
	case "":
	default:
		return nil, fmt.Errorf("%w: invalid sortBy", errInvalidRequest)
	}

	orderBy := storage.OrderType(q.Get("orderBy"))
	switch orderBy {
	case storage.AscendingOrder, storage.DescendingOrder:
		out.OrderBy = orderBy
	case "":
	default:
		return nil, fmt.Errorf("%w: invalid orderBy", errInvalidRequest)
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

	if s := q.Get("likedBy"); s != "" {
		out.LikedBy = &s
	}

	if s := q.Get("followedBy"); s != "" {
		out.FollowedBy = &s
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
	liked map[storage.PostID]community.LikeWeight,
) ListPostsResponse {
	out := ListPostsResponse{}

	out.Posts = make([]*Post, len(posts))
	for i, v := range posts {
		out.Posts[i] = toAPIPost(v)
	}

	out.Profiles = make(map[string]Profile, len(out.Profiles))
	for _, v := range profiles {
		out.Profiles[v.Address] = *toAPIProfile(v)
	}

	out.Stats = make(map[string][]StatsItem, len(stats))

	for k, v := range stats {
		out.Stats[fmt.Sprintf("%s/%s", k.Owner, k.UUID)] = toAPIStats(v)
	}

	for _, v := range out.Posts {
		l := liked[storage.PostID{Owner: v.Owner, UUID: v.UUID}]
		v.LikeWeight = &l
	}

	return out
}

func toAPIPost(p *storage.Post) *Post {
	if p == nil {
		return nil
	}

	return &Post{
		UUID:          p.UUID,
		Owner:         p.Owner,
		Title:         p.Title,
		Category:      p.Category,
		PreviewImage:  p.PreviewImage,
		Text:          p.Text,
		LikesCount:    p.Likes,
		DislikesCount: p.Dislikes,
		PDV:           utils.TokenToFloat64(sdk.NewInt(p.UPDV)),
		CreatedAt:     uint64(p.CreatedAt.Unix()),
	}
}

func toAPIProfile(p *storage.Profile) *Profile {
	if p == nil {
		return nil
	}

	return &Profile{
		Address:      p.Address,
		FirstName:    p.FirstName,
		LastName:     p.LastName,
		Bio:          p.Bio,
		Avatar:       p.Avatar,
		Gender:       p.Gender,
		Birthday:     p.Birthday,
		RegisteredAt: uint64(p.CreatedAt.Unix()),
		PostsCount:   p.PostsCount,
	}
}

func toAPIStats(s storage.Stats) []StatsItem {
	o := make([]StatsItem, 0, len(s))

	for k, v := range s {
		if k == "0001-01-01" {
			continue
		}

		o = append(o, StatsItem{
			Date:  k,
			Value: utils.TokenToFloat64(sdk.NewInt(v)),
		})
	}

	return o
}
