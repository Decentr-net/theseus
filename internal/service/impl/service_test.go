package impl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	community "github.com/Decentr-net/decentr/x/community/types"

	"github.com/Decentr-net/theseus/internal/entities"
	"github.com/Decentr-net/theseus/internal/service"
	storageinterface "github.com/Decentr-net/theseus/internal/storage"
	storage "github.com/Decentr-net/theseus/internal/storage/mock"
)

func TestSrv_OnHeight(t *testing.T) {
	tt := []struct {
		name   string
		height uint64

		err error
	}{
		{
			name:   "success",
			height: 100,
		},
		{
			name:   "fail",
			height: 100,
			err:    context.Canceled,
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			s := storage.NewMockStorage(ctrl)

			srv := New(s)

			s.EXPECT().WithLockedHeight(gomock.Any(), tc.height, gomock.Any()).DoAndReturn(func(_ context.Context, _ uint64, f func(s storageinterface.Storage) error) error {
				return f(s)
			})

			err := srv.OnHeight(context.Background(), tc.height, func(s service.Service) error {
				return tc.err
			})

			require.True(t, errors.Is(err, tc.err))
		})
	}
}

func TestSrv_GetHeight(t *testing.T) {
	ctrl := gomock.NewController(t)

	s := storage.NewMockStorage(ctrl)

	srv := New(s)

	s.EXPECT().GetHeight(gomock.Any()).Return(uint64(1), nil)
	h, err := srv.GetHeight(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, 1, h)

	s.EXPECT().GetHeight(gomock.Any()).Return(uint64(0), context.Canceled)
	h, err = srv.GetHeight(context.Background())
	require.Error(t, err)
	require.Zero(t, h)
}

func TestSrv_CreatePost(t *testing.T) {
	ctrl := gomock.NewController(t)

	s := storage.NewMockStorage(ctrl)

	srv := New(s)

	p := &entities.Post{
		UUID:         "1",
		Owner:        "2",
		Title:        "3",
		Category:     4,
		PreviewImage: "5",
		Text:         "6",
		CreatedAt:    time.Now(),
	}

	s.EXPECT().CreatePost(gomock.Any(), p).Return(nil)
	require.NoError(t, srv.CreatePost(context.Background(), p))

	s.EXPECT().CreatePost(gomock.Any(), p).Return(context.Canceled)
	require.Error(t, srv.CreatePost(context.Background(), p))
}

func TestSrv_DeletePost(t *testing.T) {
	ctrl := gomock.NewController(t)

	s := storage.NewMockStorage(ctrl)

	srv := New(s)
	timestamp := time.Now()

	s.EXPECT().DeletePost(gomock.Any(), "1", "2", timestamp, "3").Return(nil)
	require.NoError(t, srv.DeletePost(context.Background(), "1", "2", timestamp, "3"))

	s.EXPECT().DeletePost(gomock.Any(), "1", "2", timestamp, "3").Return(context.Canceled)
	require.Error(t, srv.DeletePost(context.Background(), "1", "2", timestamp, "3"))
}

func TestSrv_SetLike(t *testing.T) {
	ctrl := gomock.NewController(t)

	s := storage.NewMockStorage(ctrl)

	srv := New(s)
	timestamp := time.Now()

	s.EXPECT().SetLike(gomock.Any(), "1", "2", community.LikeWeight(1), timestamp, "3").Return(nil)
	require.NoError(t, srv.SetLike(context.Background(), "1", "2", 1, timestamp, "3"))

	s.EXPECT().SetLike(gomock.Any(), "1", "2", community.LikeWeight(1), timestamp, "3").Return(context.Canceled)
	require.Error(t, srv.SetLike(context.Background(), "1", "2", 1, timestamp, "3"))
}
