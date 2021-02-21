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

	"github.com/Decentr-net/theseus/internal/storage"
	"github.com/Decentr-net/theseus/internal/storage/mock"
)

func Test_listPosts(t *testing.T) {
	timestamp := time.Unix(100, 0)

	query := "category=1&sort_by=likes&order_by=asc&limit=100&after=1234/4321&from=1&to=1000&owner=addr&liked_by=1234"

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
			PDV:          3,
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
			PDV:          3,
		},
	}, nil)

	s.EXPECT().GetProfiles(gomock.Any(), []string{"owner", "owner2"}).Return([]*storage.Profile{
		{
			Address:   "owner",
			FirstName: "f",
			LastName:  "l",
			Avatar:    "a",
			Gender:    "g",
			Birthday:  "b",
			CreatedAt: timestamp,
		},
		{
			Address:   "owner2",
			FirstName: "f2",
			LastName:  "l2",
			Avatar:    "a2",
			Gender:    "g2",
			Birthday:  "b2",
			CreatedAt: timestamp,
		},
	}, nil)
	s.EXPECT().GetStats(gomock.Any(), []storage.PostID{
		{"owner", "uuid"}, {"owner2", "uuid2"},
	}).Return(map[storage.PostID]storage.Stats{
		{"owner", "uuid"}:   {timestamp: 1},
		{"owner2", "uuid2"}: {timestamp: 2},
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
         "preview_image":"preview",
         "text":"text",
         "likes":1,
         "dislikes":2,
         "pdv":3,
         "created_at":100
      },
      {
         "uuid":"uuid2",
         "owner":"owner2",
         "title":"title2",
         "category":2,
         "preview_image":"preview2",
         "text":"text2",
         "likes":1,
         "dislikes":2,
         "pdv":3,
         "created_at":100
      }
   ],
   "profiles":{
      "owner":{
         "address":"owner",
         "first_name":"f",
         "last_name":"l",
         "avatar":"a",
         "gender":"g",
         "birthday":"b",
         "created_at":100
      },
      "owner2":{
         "address":"owner2",
         "first_name":"f2",
         "last_name":"l2",
         "avatar":"a2",
         "gender":"g2",
         "birthday":"b2",
         "created_at":100
      }
   },
   "stats":{
      "owner/uuid":{
         "1970-01-01":1
      },
      "owner2/uuid2":{
         "1970-01-01":2
      }
   }
}
	`, w.Body.String())
}

func Test_getPost(t *testing.T) {
	timestamp := time.Unix(3000, 0)

	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/posts/%s", "owner/uuid"), nil)
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
		PDV:          3,
	}, nil)

	router := chi.NewRouter()
	s := server{s: srv}
	router.Get("/v1/posts/{owner}/{uuid}", s.getPost)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `
{
   "uuid":"owner",
   "owner":"owner",
   "title":"title",
   "category":1,
   "preview_image":"preview",
   "text":"text",
   "likes":1,
   "dislikes":2,
   "pdv":3,
   "created_at":3000
}
`, w.Body.String())
}
