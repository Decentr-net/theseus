package blockchain

import (
	"context"
	"errors"
	"testing"
	"time"

	ctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Decentr-net/ariadne"
	ariadnemock "github.com/Decentr-net/ariadne/mock"
	communitytypes "github.com/Decentr-net/decentr/x/community/types"
	operationstypes "github.com/Decentr-net/decentr/x/operations/types"

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
			msg: &communitytypes.MsgCreatePost{
				Post: communitytypes.Post{
					Uuid:         "1234",
					Owner:        owner.String(),
					Title:        "title",
					Category:     communitytypes.Category_CATEGORY_WORLD_NEWS,
					PreviewImage: "url",
					Text:         "text",
				},
			},
			expect: func(s *storagemock.MockStorage) {
				s.EXPECT().CreatePost(gomock.Any(), &storage.CreatePostParams{
					UUID:         "1234",
					Owner:        "decentr1u9slwz3sje8j94ccpwlslflg0506yc8y2ylmtz",
					Title:        "title",
					Category:     communitytypes.Category_CATEGORY_WORLD_NEWS,
					PreviewImage: "url",
					Text:         "text",
					CreatedAt:    timestamp,
				})
			},
		},
		{
			name: "like_post",
			msg: &communitytypes.MsgSetLike{
				Like: communitytypes.Like{
					PostOwner: owner.String(),
					PostUuid:  "1234",
					Owner:     owner.String(),
					Weight:    communitytypes.LikeWeight_LIKE_WEIGHT_DOWN,
				},
			},
			expect: func(s *storagemock.MockStorage) {
				// nolint
				s.EXPECT().GetLikes(gomock.Any(), owner.String(), storage.PostID{
					Owner: owner.String(),
					UUID:  "1234",
				}).Return(map[storage.PostID]communitytypes.LikeWeight{
					storage.PostID{
						Owner: owner.String(),
						UUID:  "1234",
					}: communitytypes.LikeWeight_LIKE_WEIGHT_UP,
				}, nil)

				s.EXPECT().AddPDV(gomock.Any(), owner.String(), int64(-2), timestamp).Return(nil)

				s.EXPECT().GetPost(gomock.Any(), storage.PostID{Owner: "decentr1u9slwz3sje8j94ccpwlslflg0506yc8y2ylmtz", UUID: "1234"}).Return(&storage.Post{}, nil)

				s.EXPECT().SetLike(
					gomock.Any(),
					storage.PostID{Owner: "decentr1u9slwz3sje8j94ccpwlslflg0506yc8y2ylmtz", UUID: "1234"},
					communitytypes.LikeWeight_LIKE_WEIGHT_DOWN,
					timestamp,
					"decentr1u9slwz3sje8j94ccpwlslflg0506yc8y2ylmtz",
				)
			},
		},
		{
			name: "delete_post",
			msg: &communitytypes.MsgDeletePost{
				PostOwner: owner.String(),
				PostUuid:  "1234",
				Owner:     owner.String(),
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
			name: "follow",
			msg: &communitytypes.MsgFollow{
				Owner: owner.String(),
				Whom:  owner2.String(),
			},
			expect: func(s *storagemock.MockStorage) {
				s.EXPECT().Follow(gomock.Any(), owner.String(), owner2.String())
			},
		},
		{
			name: "unfollow",
			msg: &communitytypes.MsgUnfollow{
				Owner: owner.String(),
				Whom:  owner2.String(),
			},
			expect: func(s *storagemock.MockStorage) {
				s.EXPECT().Unfollow(gomock.Any(), owner.String(), owner2.String())
			},
		},
		{
			name: "distribute_rewards",
			msg: &operationstypes.MsgDistributeRewards{
				Owner: owner.String(),
				Rewards: []operationstypes.Reward{
					{
						Receiver: owner.String(),
						Reward:   sdk.DecProto{Dec: sdk.NewDecWithPrec(100, 6)},
					},
					{
						Receiver: owner2.String(),
						Reward:   sdk.DecProto{Dec: sdk.NewDecWithPrec(10, 6)},
					},
				},
			},
			expect: func(s *storagemock.MockStorage) {
				s.EXPECT().AddPDV(gomock.Any(), owner.String(), int64(100), timestamp)
				s.EXPECT().AddPDV(gomock.Any(), owner2.String(), int64(10), timestamp)
			},
		},
		{
			name: "wipe_account",
			msg: &operationstypes.MsgResetAccount{
				Owner:   owner.String(),
				Address: owner.String(),
			},
			expect: func(s *storagemock.MockStorage) {
				s.EXPECT().ResetAccount(gomock.Any(), owner.String())
			},
		},
	}

	for i := range tt {
		tc := tt[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := storagemock.NewMockStorage(gomock.NewController(t))

			s.EXPECT().InTx(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, f func(_ storage.Storage) error) error {
				return f(s)
			})
			s.EXPECT().SetHeight(gomock.Any(), uint64(1)).Return(nil)
			s.EXPECT().RefreshViews(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			tc.expect(s)

			msg, err := ctypes.NewAnyWithValue(tc.msg)
			require.NoError(t, err)
			block := ariadne.Block{
				Height: 1,
				Time:   timestamp,
				Txs: []sdk.Tx{
					&tx.Tx{
						Body: &tx.TxBody{
							Messages: []*ctypes.Any{
								msg,
							},
						},
					},
				},
			}

			require.NoError(t, blockchain{s: s}.processBlockFunc(context.Background())(block))
		})
	}
}

func TestBlockchain_processBlockFunc_errors(t *testing.T) {
	s := storagemock.NewMockStorage(gomock.NewController(t))

	s.EXPECT().InTx(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, f func(_ storage.Storage) error) error {
		return context.Canceled
	})

	require.Error(t, blockchain{s: s}.processBlockFunc(context.Background())(ariadne.Block{Height: 1}))
}
