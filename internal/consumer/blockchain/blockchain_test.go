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
	pdv "github.com/Decentr-net/decentr/x/pdv/types"
	profile "github.com/Decentr-net/decentr/x/profile/types"

	"github.com/Decentr-net/theseus/internal/storage"
	storagemock "github.com/Decentr-net/theseus/internal/storage/mock"
)

var errTest = errors.New("test")

func TestBlockchain_Run(t *testing.T) {
	ctrl := gomock.NewController(t)

	f, s := ariadnemock.NewMockFetcher(ctrl), storagemock.NewMockStorage(ctrl)

	b := New(f, s, time.Nanosecond, time.Nanosecond)

	s.EXPECT().GetHeight(gomock.Any()).Return(uint64(1), nil)

	f.EXPECT().FetchBlocks(gomock.Any(), uint64(1), gomock.Any(), gomock.Any()).Return(nil)

	require.NoError(t, b.Run(context.Background()))
}

func TestBlockchain_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)

	f, s := ariadnemock.NewMockFetcher(ctrl), storagemock.NewMockStorage(ctrl)

	b := New(f, s, time.Nanosecond, time.Nanosecond)

	s.EXPECT().GetHeight(gomock.Any()).Return(uint64(1), nil)

	f.EXPECT().FetchBlocks(gomock.Any(), uint64(1), gomock.Any(), gomock.Any()).Return(errTest)

	require.Equal(t, errTest, b.Run(context.Background()))
}

func TestBlockchain_processBlockFunc(t *testing.T) {
	timestamp := time.Now()
	owner, err := sdk.AccAddressFromBech32("decentr1u9slwz3sje8j94ccpwlslflg0506yc8y2ylmtz")
	require.NoError(t, err)

	owner2, err := sdk.AccAddressFromBech32("decentr1ltx6yymrs8eq4nmnhzfzxj6tspjuymh8mgd6gz")
	require.NoError(t, err)

	tt := []struct {
		name   string
		msg    sdk.Msg
		expect func(s *storagemock.MockStorage)
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
			expect: func(s *storagemock.MockStorage) {
				s.EXPECT().CreatePost(gomock.Any(), &storage.CreatePostParams{
					UUID:         "1234",
					Owner:        "decentr1u9slwz3sje8j94ccpwlslflg0506yc8y2ylmtz",
					Title:        "title",
					Category:     community.WorldNewsCategory,
					PreviewImage: "url",
					Text:         "text",
					CreatedAt:    timestamp,
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
			expect: func(s *storagemock.MockStorage) {
				// nolint
				s.EXPECT().GetLikes(gomock.Any(), owner.String(), storage.PostID{
					Owner: owner.String(),
					UUID:  "1234",
				}).Return(map[storage.PostID]community.LikeWeight{
					storage.PostID{
						Owner: owner.String(),
						UUID:  "1234",
					}: community.LikeWeightUp,
				}, nil)

				s.EXPECT().AddPDV(gomock.Any(), owner.String(), int64(-2), timestamp).Return(nil)

				s.EXPECT().SetLike(
					gomock.Any(),
					storage.PostID{Owner: "decentr1u9slwz3sje8j94ccpwlslflg0506yc8y2ylmtz", UUID: "1234"},
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
			expect: func(s *storagemock.MockStorage) {
				s.EXPECT().DeletePost(gomock.Any(),
					storage.PostID{Owner: "decentr1u9slwz3sje8j94ccpwlslflg0506yc8y2ylmtz", UUID: "1234"},
					timestamp,
					"decentr1u9slwz3sje8j94ccpwlslflg0506yc8y2ylmtz",
				)
			},
		},
		{
			name: "set_profile",
			msg: profile.MsgSetPublic{
				Owner: owner,
				Public: profile.Public{
					FirstName: "first_name",
					LastName:  "last_name",
					Bio:       "bio",
					Avatar:    "avatar",
					Gender:    "male",
					Birthday:  "01.02.2006",
				},
			},
			expect: func(s *storagemock.MockStorage) {
				s.EXPECT().GetProfileStats(gomock.Any(), owner.String()).Return(storage.Stats{}, nil)
				s.EXPECT().SetProfile(gomock.Any(), &storage.SetProfileParams{
					Address:   owner.String(),
					FirstName: "first_name",
					LastName:  "last_name",
					Bio:       "bio",
					Avatar:    "avatar",
					Gender:    "male",
					Birthday:  "01.02.2006",
					CreatedAt: timestamp,
				})
			},
		},
		{
			name: "set_profile_new",
			msg: profile.MsgSetPublic{
				Owner: owner,
				Public: profile.Public{
					FirstName: "first_name",
					LastName:  "last_name",
					Bio:       "bio",
					Avatar:    "avatar",
					Gender:    "male",
					Birthday:  "01.02.2006",
				},
			},
			expect: func(s *storagemock.MockStorage) {
				s.EXPECT().GetProfileStats(gomock.Any(), owner.String()).Return(nil, storage.ErrNotFound)
				s.EXPECT().AddPDV(gomock.Any(), owner.String(), int64(1000000), timestamp).Return(nil)
				s.EXPECT().SetProfile(gomock.Any(), &storage.SetProfileParams{
					Address:   owner.String(),
					FirstName: "first_name",
					LastName:  "last_name",
					Bio:       "bio",
					Avatar:    "avatar",
					Gender:    "male",
					Birthday:  "01.02.2006",
					CreatedAt: timestamp,
				})
			},
		},
		{
			name: "follow",
			msg: community.MsgFollow{
				Owner: owner,
				Whom:  owner2,
			},
			expect: func(s *storagemock.MockStorage) {
				s.EXPECT().Follow(gomock.Any(), owner.String(), owner2.String())
			},
		},
		{
			name: "unfollow",
			msg: community.MsgUnfollow{
				Owner: owner,
				Whom:  owner2,
			},
			expect: func(s *storagemock.MockStorage) {
				s.EXPECT().Unfollow(gomock.Any(), owner.String(), owner2.String())
			},
		},
		{
			name: "distribute_rewards",
			msg: pdv.MsgDistributeRewards{
				Owner: owner,
				Rewards: []pdv.Reward{
					{
						Receiver: owner,
						Reward:   100,
					},
					{
						Receiver: owner2,
						Reward:   10,
					},
				},
			},
			expect: func(s *storagemock.MockStorage) {
				s.EXPECT().AddPDV(gomock.Any(), owner.String(), int64(100), timestamp)
				s.EXPECT().AddPDV(gomock.Any(), owner2.String(), int64(10), timestamp)
			},
		},
	}

	for i := range tt {
		tc := tt[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := storagemock.NewMockStorage(gomock.NewController(t))

			s.EXPECT().WithLockedHeight(gomock.Any(), uint64(1), gomock.Any()).DoAndReturn(func(_ context.Context, _ uint64, f func(_ storage.Storage) error) error {
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
	s := storagemock.NewMockStorage(gomock.NewController(t))

	s.EXPECT().WithLockedHeight(gomock.Any(), uint64(1), gomock.Any()).DoAndReturn(func(_ context.Context, _ uint64, f func(_ storage.Storage) error) error {
		return context.Canceled
	})

	require.Error(t, blockchain{s: s}.processBlockFunc(context.Background())(ariadne.Block{Height: 1}))

	s.EXPECT().WithLockedHeight(gomock.Any(), uint64(1), gomock.Any()).DoAndReturn(func(_ context.Context, _ uint64, f func(_ storage.Storage) error) error {
		return storage.ErrRequestedHeightIsTooLow
	})

	require.NoError(t, blockchain{s: s}.processBlockFunc(context.Background())(ariadne.Block{Height: 1}))
}
