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
			Likes:        1,
			Dislikes:     2,
			UPDV:         3,
		},
	}, nil)

	s.EXPECT().GetProfiles(gomock.Any(), "owner", "owner2").Return([]*storage.Profile{
		{
			Address:    "owner",
			FirstName:  "f",
			LastName:   "l",
			Bio:        "b",
			Avatar:     "a",
			Gender:     "g",
			Birthday:   "b",
			CreatedAt:  timestamp,
			PostsCount: 1,
		},
		{
			Address:    "owner2",
			FirstName:  "f2",
			LastName:   "l2",
			Bio:        "b2",
			Avatar:     "a2",
			Gender:     "g2",
			Birthday:   "b2",
			CreatedAt:  timestamp,
			PostsCount: 4,
		},
	}, nil)

	s.EXPECT().GetStats(
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
		 "likeWeight": 1,
         "createdAt":100
      }
   ],
   "profiles":{
      "owner":{
         "address":"owner",
         "firstName":"f",
         "lastName":"l",
		 "bio":"b",
         "avatar":"a",
         "gender":"g",
         "birthday":"b",
         "registeredAt":100,
		 "postsCount": 1
      },
      "owner2":{
         "address":"owner2",
         "firstName":"f2",
         "lastName":"l2",
		 "bio":"b2",
         "avatar":"a2",
         "gender":"g2",
         "birthday":"b2",
         "registeredAt":100,
		 "postsCount": 4
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
		Likes:        1,
		Dislikes:     2,
		UPDV:         3,
	}, nil)

	srv.EXPECT().GetProfiles(gomock.Any(), "owner").Return([]*storage.Profile{
		{
			Address:    "owner",
			FirstName:  "f",
			LastName:   "l",
			Bio:        "b",
			Avatar:     "a",
			Gender:     "g",
			Birthday:   "b",
			CreatedAt:  timestamp,
			PostsCount: 0,
		},
	}, nil)

	srv.EXPECT().GetStats(
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
		"pdv":3e-6,
		"likeWeight": -1,
		"createdAt":3000
	},
    "profile":{
		"address":"owner",
		"firstName":"f",
		"lastName":"l",
		"bio":"b",
		"avatar":"a",
		"gender":"g",
		"birthday":"b",
		"registeredAt":3000,
		"postsCount":0
	},
	"stats": [
		{ "date":"1970-01-01", "value":1e-6 }
	]
}
`, w.Body.String())
}

func Test_getProfileStats(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/v1/profiles/owner/stats", nil)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	srv := mock.NewMockStorage(ctrl)

	srv.EXPECT().GetProfileStats(gomock.Any(), "owner").Return(storage.Stats{
		"1970-01-01": 1,
	}, nil)

	router := chi.NewRouter()
	s := server{s: srv}
	router.Get("/v1/profiles/{address}/stats", s.getProfileStats)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `[{ "date":"1970-01-01", "value":1e-6 }]`, w.Body.String())
}

func Test_getProfileStats_not_found(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/v1/profiles/owner/stats", nil)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	srv := mock.NewMockStorage(ctrl)

	srv.EXPECT().GetProfileStats(gomock.Any(), "owner").Return(nil, storage.ErrNotFound)

	router := chi.NewRouter()
	s := server{s: srv}
	router.Get("/v1/profiles/{address}/stats", s.getProfileStats)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `[]`, w.Body.String())
}

func Test_getAllUsersStats(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/v1/profiles/stats", nil)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	srv := mock.NewMockStorage(ctrl)

	srv.EXPECT().GetDecentrStats(gomock.Any()).Return(&storage.DecentrStats{
		ADV: 1.01,
		DDV: 2,
	}, nil)

	router := chi.NewRouter()
	s := server{s: srv}
	router.Get("/v1/profiles/stats", s.getAllUsersStats)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{ "adv":1.01, "ddv":2 }`, w.Body.String())
}
