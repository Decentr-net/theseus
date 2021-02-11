package blockchain

import (
	"context"
	"errors"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Decentr-net/ariadne"
	ariadnemock "github.com/Decentr-net/ariadne/mock"
	community "github.com/Decentr-net/decentr/x/community/types"

	"github.com/Decentr-net/theseus/internal/entities"
	"github.com/Decentr-net/theseus/internal/service"
	servicemock "github.com/Decentr-net/theseus/internal/service/mock"
)

var errTest = errors.New("test")

func TestBlockchain_Run(t *testing.T) {
	ctrl := gomock.NewController(t)

	f, s := ariadnemock.NewMockFetcher(ctrl), servicemock.NewMockService(ctrl)

	b := New(f, s, time.Nanosecond, time.Nanosecond)

	s.EXPECT().GetHeight(gomock.Any()).Return(uint64(1), nil)

	f.EXPECT().FetchBlocks(gomock.Any(), uint64(1), gomock.Any(), gomock.Any()).Return(nil)

	require.NoError(t, b.Run(context.Background()))
}

func TestBlockchain_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)

	f, s := ariadnemock.NewMockFetcher(ctrl), servicemock.NewMockService(ctrl)

	b := New(f, s, time.Nanosecond, time.Nanosecond)

	s.EXPECT().GetHeight(gomock.Any()).Return(uint64(1), nil)

	f.EXPECT().FetchBlocks(gomock.Any(), uint64(1), gomock.Any(), gomock.Any()).Return(errTest)

	require.Equal(t, errTest, b.Run(context.Background()))
}

func TestBlockchain_processBlockFunc(t *testing.T) {
	timestamp := time.Now()
	owner, err := sdk.AccAddressFromBech32("decentr1u9slwz3sje8j94ccpwlslflg0506yc8y2ylmtz")
	require.NoError(t, err)

	tt := []struct {
		name   string
		msg    sdk.Msg
		expect func(s *servicemock.MockService)
	}{
		{
			name: "create_post",
			msg: community.MsgCreatePost{
				UUID:         "1234",
				Owner:        owner,
				Title:        "title",
				Category:     community.WorldNewsCategory,
				PreviewImage: "url",
				Text:         "text",
			},
			expect: func(s *servicemock.MockService) {
				s.EXPECT().CreatePost(gomock.Any(), &entities.Post{
					UUID:         "1234",
					Owner:        "decentr1u9slwz3sje8j94ccpwlslflg0506yc8y2ylmtz",
					Title:        "title",
					Category:     community.WorldNewsCategory,
					PreviewImage: "url",
					Text:         "text",
					CreatedAt:    timestamp,
					DeletedAt:    nil,
					DeletedBy:    nil,
				})
			},
		},
		{
			name: "like_post",
			msg: community.MsgSetLike{
				PostOwner: owner,
				PostUUID:  "1234",
				Owner:     owner,
				Weight:    community.LikeWeightDown,
			},
			expect: func(s *servicemock.MockService) {
				s.EXPECT().SetLike(
					gomock.Any(),
					"decentr1u9slwz3sje8j94ccpwlslflg0506yc8y2ylmtz",
					"1234",
					community.LikeWeightDown,
					timestamp,
					"decentr1u9slwz3sje8j94ccpwlslflg0506yc8y2ylmtz",
				)
			},
		},
		{
			name: "delete_post",
			msg: community.MsgDeletePost{
				PostOwner: owner,
				PostUUID:  "1234",
				Owner:     owner,
			},
			expect: func(s *servicemock.MockService) {
				s.EXPECT().DeletePost(gomock.Any(),
					"decentr1u9slwz3sje8j94ccpwlslflg0506yc8y2ylmtz",
					"1234",
					timestamp,
					"decentr1u9slwz3sje8j94ccpwlslflg0506yc8y2ylmtz",
				)
			},
		},
	}

	for i := range tt {
		tc := tt[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := servicemock.NewMockService(gomock.NewController(t))

			s.EXPECT().OnHeight(gomock.Any(), uint64(1), gomock.Any()).DoAndReturn(func(_ context.Context, _ uint64, f func(_ service.Service) error) error {
				return f(s)
			})
			tc.expect(s)

			block := ariadne.Block{
				Height: 1,
				Time:   timestamp,
				Txs: []sdk.Tx{
					auth.StdTx{
						Msgs: []sdk.Msg{tc.msg},
					},
				},
			}

			require.NoError(t, blockchain{s: s}.processBlockFunc(context.Background())(block))
		})
	}
}

func TestBlockchain_processBlockFunc_errors(t *testing.T) {
	s := servicemock.NewMockService(gomock.NewController(t))

	s.EXPECT().OnHeight(gomock.Any(), uint64(1), gomock.Any()).DoAndReturn(func(_ context.Context, _ uint64, f func(_ service.Service) error) error {
		return service.ErrRequestedHeightIsTooHigh
	})

	require.Error(t, blockchain{s: s}.processBlockFunc(context.Background())(ariadne.Block{Height: 1}))

	s.EXPECT().OnHeight(gomock.Any(), uint64(1), gomock.Any()).DoAndReturn(func(_ context.Context, _ uint64, f func(_ service.Service) error) error {
		return service.ErrRequestedHeightIsTooLow
	})

	require.NoError(t, blockchain{s: s}.processBlockFunc(context.Background())(ariadne.Block{Height: 1}))
}
