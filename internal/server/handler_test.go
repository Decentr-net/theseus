package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	community "github.com/Decentr-net/decentr/x/community/types"

	"github.com/Decentr-net/theseus/internal/storage"
	"github.com/Decentr-net/theseus/internal/storage/mock"
)

func Test_listPosts(t *testing.T) {
	timestamp := time.Unix(100, 0)

	query := "category=1&sortBy=likesCount&orderBy=asc&limit=100&after=1234/4321&from=1&to=1000&owner=addr&likedBy=1234&followedBy=111&requestedBy=owner"

	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/posts?%s", query), nil)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mock.NewMockStorage(ctrl)

	s.EXPECT().ListPosts(gomock.Any(), gomock.Any()).Do(func(_ context.Context, p *storage.ListPostsParams) {
		assert.Equal(t, storage.LikesSortType, p.SortBy)
		assert.Equal(t, storage.AscendingOrder, p.OrderBy)
		assert.EqualValues(t, 1, *p.Category)
		assert.Equal(t, "addr", *p.Owner)
		assert.Equal(t, "1234", *p.LikedBy)
		assert.Equal(t, "111", *p.FollowedBy)
		assert.EqualValues(t, 100, p.Limit)
		assert.Equal(t, storage.PostID{
			Owner: "1234",
			UUID:  "4321",
		}, *p.After)
		assert.EqualValues(t, 1, *p.From)
		assert.EqualValues(t, 1000, *p.To)
	}).Return([]*storage.Post{
		{
			UUID:         "uuid",
			Owner:        "owner",
			Title:        "title",
			Category:     1,
			PreviewImage: "preview",
			Text:         "text",
			CreatedAt:    timestamp,
			Slug:         "slug1",
			Likes:        1,
			Dislikes:     2,
			UPDV:         3,
		},
		{
			UUID:         "uuid2",
			Owner:        "owner2",
			Title:        "title2",
			Category:     2,
			PreviewImage: "preview2",
			Text:         "text2",
			CreatedAt:    timestamp,
			Slug:         "slug2",
			Likes:        1,
			Dislikes:     2,
			UPDV:         3,
		},
	}, nil)

	s.EXPECT().GetProfileStats(gomock.Any(), "owner", "owner2").Return([]*storage.ProfileStats{
		{
			Address:    "owner",
			PostsCount: 1,
			Stats:      storage.Stats{"0001-01-01": 1, "1970-01-01": 2},
		},
		{
			Address:    "owner2",
			PostsCount: 4,
			Stats:      storage.Stats{"1970-01-02": 1},
		},
	}, nil)

	s.EXPECT().GetPostStats(
		gomock.Any(),
		storage.PostID{"owner", "uuid"},
		storage.PostID{"owner2", "uuid2"},
	).Return(map[storage.PostID]storage.Stats{
		{"owner", "uuid"}:   {"1970-01-01": 1},
		{"owner2", "uuid2"}: {"1970-01-01": 2},
	}, nil)

	s.EXPECT().GetLikes(
		gomock.Any(),
		"owner",
		storage.PostID{"owner", "uuid"},
		storage.PostID{"owner2", "uuid2"},
	).Return(map[storage.PostID]community.LikeWeight{
		{"owner", "uuid"}:   0,
		{"owner2", "uuid2"}: 1,
	}, nil)

	router := chi.NewRouter()
	srv := server{s: s}
	router.Get("/v1/posts", srv.listPosts)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `
{
   "posts":[
      {
         "uuid":"uuid",
         "owner":"owner",
         "title":"title",
         "category":1,
         "previewImage":"preview",
         "text":"text",
         "likesCount":1,
         "dislikesCount":2,
         "pdv":3e-6,
         "slug": "slug1",
         "likeWeight": 0,
         "createdAt":100
      },
      {
         "uuid":"uuid2",
         "owner":"owner2",
         "title":"title2",
         "category":2,
         "previewImage":"preview2",
         "text":"text2",
         "likesCount":1,
         "dislikesCount":2,
         "pdv":3e-6,
         "slug": "slug2",
         "likeWeight": 1,
         "createdAt":100
      }
   ],
   "profileStats":{
      "owner":{
		 "postsCount": 1,
		 "stats": [{ "date": "1970-01-01", "value": 1.000002 }]
      },
      "owner2":{
		 "postsCount": 4,
		 "stats": [{ "date": "1970-01-02", "value": 1.000001 }]
      }
   },
   "stats":{
      "owner/uuid":[
         { "date": "1970-01-01", "value": 1e-6 }
      ],
      "owner2/uuid2": [
         { "date": "1970-01-01", "value": 2e-6 }
      ]
   }
}
	`, w.Body.String())
}

func Test_getPost(t *testing.T) {
	timestamp := time.Unix(3000, 0)

	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/posts/%s?requestedBy=owner", "owner/uuid"), nil)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	srv := mock.NewMockStorage(ctrl)

	srv.EXPECT().GetPost(gomock.Any(), storage.PostID{
		Owner: "owner",
		UUID:  "uuid",
	}).Return(&storage.Post{
		UUID:         "uuid",
		Owner:        "owner",
		Title:        "title",
		Category:     1,
		PreviewImage: "preview",
		Text:         "text",
		CreatedAt:    timestamp,
		Slug:         "slug",
		Likes:        1,
		Dislikes:     2,
		UPDV:         3,
	}, nil)

	srv.EXPECT().GetProfileStats(gomock.Any(), "owner").Return([]*storage.ProfileStats{
		{
			Address:    "owner",
			PostsCount: 0,
			Stats:      storage.Stats{},
		},
	}, nil)

	srv.EXPECT().GetPostStats(
		gomock.Any(),
		storage.PostID{"owner", "uuid"},
	).Return(map[storage.PostID]storage.Stats{
		{"owner", "uuid"}: {"1970-01-01": 1},
	}, nil)

	srv.EXPECT().GetLikes(
		gomock.Any(),
		"owner",
		storage.PostID{"owner", "uuid"},
	).Return(map[storage.PostID]community.LikeWeight{
		{"owner", "uuid"}: -1,
	}, nil)

	router := chi.NewRouter()
	s := server{s: srv}
	router.Get("/v1/posts/{owner}/{uuid}", s.getPost)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `
{
	"post": {
		"uuid":"uuid",
		"owner":"owner",
		"title":"title",
		"category":1,
		"previewImage":"preview",
		"text":"text",
		"likesCount":1,
		"dislikesCount":2,
		"slug": "slug",
		"pdv":3e-6,
		"likeWeight": -1,
		"createdAt":3000
	},
    "profileStats":{
		"postsCount":0,
		"stats": []
	},
	"stats": [
		{ "date":"1970-01-01", "value":1e-6 }
	]
}
`, w.Body.String())
}

func Test_getSharePostBySlug(t *testing.T) {
	timestamp := time.Unix(3000, 0)

	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/posts/%s", "slug"), nil)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	srv := mock.NewMockStorage(ctrl)

	srv.EXPECT().GetPostBySlug(gomock.Any(), "slug").Return(&storage.Post{
		UUID:         "uuid",
		Owner:        "owner",
		Title:        "title",
		Category:     1,
		PreviewImage: "preview",
		Text:         "text",
		CreatedAt:    timestamp,
		Slug:         "slug",
		Likes:        1,
		Dislikes:     2,
		UPDV:         3,
	}, nil)

	router := chi.NewRouter()
	s := server{s: srv}
	router.Get("/v1/posts/{slug}", s.getSharePostBySlug)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.JSONEq(t, `
	{
		"uuid":"uuid",
		"owner":"owner",
		"title":"title"
	}
`, w.Body.String())
}

func Test_getProfileStats(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/v1/profiles/owner/stats", nil)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	srv := mock.NewMockStorage(ctrl)

	srv.EXPECT().GetProfileStats(gomock.Any(), "owner").Return([]*storage.ProfileStats{
		{
			PostsCount: 1,
			Stats:      storage.Stats{"1970-01-01": 1},
		},
	}, nil)

	router := chi.NewRouter()
	s := server{s: srv}
	router.Get("/v1/profiles/{address}/stats", s.getProfileStats)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"postsCount": 1, "stats":[{ "date":"1970-01-01", "value":1.000001 }]}`, w.Body.String())
}

func Test_getDecentrStats(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/v1/profiles/stats", nil)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	srv := mock.NewMockStorage(ctrl)

	srv.EXPECT().GetDecentrStats(gomock.Any()).Return(&storage.DecentrStats{
		ADV: 1010000,
		DDV: 2000000,
	}, nil)

	router := chi.NewRouter()
	s := server{s: srv}
	router.Get("/v1/profiles/stats", s.getDecentrStats)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{ "adv":1.01, "ddv":2 }`, w.Body.String())
}
