// Package blockchain is a consumer interface.
package blockchain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Decentr-net/ariadne"
	"github.com/Decentr-net/decentr/config"
	communitytypes "github.com/Decentr-net/decentr/x/community/types"
	operationstypes "github.com/Decentr-net/decentr/x/operations/types"

	"github.com/Decentr-net/theseus/internal/consumer"
	"github.com/Decentr-net/theseus/internal/storage"
)

// nolint:gochecknoinits
func init() {
	config.SetAddressPrefixes()
}

var log = logrus.WithField("package", "blockchain")

type blockchain struct {
	f ariadne.Fetcher
	s storage.Storage

	retryInterval          time.Duration
	retryLastBlockInterval time.Duration
}

// New returns new blockchain instance.
func New(f ariadne.Fetcher, s storage.Storage, retryInterval, retryLastBlockInterval time.Duration) consumer.Consumer {
	return blockchain{
		f: f,
		s: s,

		retryInterval:          retryInterval,
		retryLastBlockInterval: retryLastBlockInterval,
	}
}

func logError(h uint64, err error) {
	log.WithField("height", h).WithError(err).Error("failed to process block")
}

func (b blockchain) Name() string {
	return "blockchain"
}

func (b blockchain) Ping(ctx context.Context) (interface{}, error) {
	return b.s.GetHeight(ctx)
}

func (b blockchain) Run(ctx context.Context) error {
	from, err := b.s.GetHeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current height: %w", err)
	}

	return b.f.FetchBlocks(ctx, from, b.processBlockFunc(ctx),
		ariadne.WithErrHandler(logError),
		ariadne.WithSkipError(false),
		ariadne.WithRetryInterval(b.retryInterval),
		ariadne.WithRetryLastBlockInterval(b.retryLastBlockInterval),
	)
}

func (b blockchain) processBlockFunc(ctx context.Context) func(block ariadne.Block) error {
	return func(block ariadne.Block) error {
		return b.s.InTx(ctx, func(s storage.Storage) error {
			log := log.WithField("height", block.Height).WithField("txs", len(block.Txs))
			log.Info("processing block")
			log.WithField("msgs", fmt.Sprintf("%+v", block.Messages())).Debug()

			needRefreshPostsView := false
			needRefreshStatsView := block.Height%50 == 0

			for _, msg := range block.Messages() {
				var err error

				switch msg := msg.(type) {
				case *communitytypes.MsgCreatePost:
					needRefreshPostsView = true
					err = processMsgCreatePost(ctx, s, block.Time, msg)
				case *communitytypes.MsgDeletePost:
					needRefreshPostsView = true
					err = processMsgDeletePost(ctx, s, block.Time, *msg)
				case *communitytypes.MsgSetLike:
					needRefreshPostsView = true
					needRefreshStatsView = true
					err = processMsgSetLike(ctx, s, block.Time, *msg)
				case *communitytypes.MsgFollow:
					err = processMsgFollow(ctx, s, *msg)
				case *communitytypes.MsgUnfollow:
					err = processMsgUnfollow(ctx, s, *msg)
				case *operationstypes.MsgDistributeRewards:
					err = processDistributeRewards(ctx, s, block.Time, msg)
				case *operationstypes.MsgResetAccount:
					err = processMsgResetAccount(ctx, s, msg.Address)
				default:
					log.WithField("msg", msg).Debug("skip message")
				}

				if err != nil {
					return fmt.Errorf("failed to process msg: %w", err)
				}
			}

			if err := s.SetHeight(ctx, block.Height); err != nil {
				return fmt.Errorf("failed to set height: %w", err)
			}

			if err := s.RefreshViews(ctx, needRefreshPostsView, needRefreshStatsView); err != nil {
				return fmt.Errorf("failed to refresh views: %w", err)
			}

			return nil
		})
	}
}

func processMsgCreatePost(ctx context.Context, s storage.Storage, timestamp time.Time, msg *communitytypes.MsgCreatePost) error {
	return s.CreatePost(ctx, &storage.CreatePostParams{
		UUID:         msg.Post.Uuid,
		Owner:        msg.Post.Owner,
		Title:        msg.Post.Title,
		Category:     msg.Post.Category,
		PreviewImage: msg.Post.PreviewImage,
		Text:         msg.Post.Text,
		CreatedAt:    timestamp,
	})
}

func processMsgDeletePost(ctx context.Context, s storage.Storage, timestamp time.Time, msg communitytypes.MsgDeletePost) error {
	if err := s.DeletePost(ctx, storage.PostID{Owner: msg.PostOwner, UUID: msg.PostUuid}, timestamp, msg.Owner); err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return err
		}
	}
	return nil
}

func processMsgSetLike(ctx context.Context, s storage.Storage, timestamp time.Time, msg communitytypes.MsgSetLike) error {
	p := storage.PostID{
		Owner: msg.Like.PostOwner,
		UUID:  msg.Like.PostUuid,
	}

	m, err := s.GetLikes(ctx, msg.Like.Owner, p)
	if err != nil {
		return fmt.Errorf("failed to get like: %w", err)
	}

	previousWeight := communitytypes.LikeWeight_LIKE_WEIGHT_ZERO
	if l, ok := m[p]; ok {
		previousWeight = l
	}

	if err := s.AddPDV(ctx, msg.Like.PostOwner, int64(msg.Like.Weight-previousWeight), timestamp); err != nil {
		return fmt.Errorf("failed to add pdv to profile stats: %w", err)
	}

	postID := storage.PostID{Owner: msg.Like.PostOwner, UUID: msg.Like.PostUuid}

	// check the related post exists
	if _, err := s.GetPost(ctx, postID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.WithField("msg", msg).Errorf("set like: related post not found")
			return nil
		}
		return err
	}

	return s.SetLike(ctx, postID, msg.Like.Weight, timestamp, msg.Like.Owner)
}

func processMsgFollow(ctx context.Context, s storage.Storage, msg communitytypes.MsgFollow) error {
	return s.Follow(ctx, msg.Owner, msg.Whom)
}

func processMsgUnfollow(ctx context.Context, s storage.Storage, msg communitytypes.MsgUnfollow) error {
	return s.Unfollow(ctx, msg.Owner, msg.Whom)
}

func processDistributeRewards(ctx context.Context, s storage.Storage, timestamp time.Time, msg *operationstypes.MsgDistributeRewards) error {
	for _, v := range msg.Rewards {
		if err := s.AddPDV(ctx, v.Receiver, v.Reward.Dec.MulInt64(storage.PDVDenominator).TruncateInt64(), timestamp); err != nil {
			return fmt.Errorf("failed to add pdv: %w", err)
		}
	}

	return nil
}

func processMsgResetAccount(ctx context.Context, s storage.Storage, owner string) error {
	if err := s.ResetAccount(ctx, owner); err != nil {
		return fmt.Errorf("failed to wipe account: %w", err)
	}

	return nil
}
