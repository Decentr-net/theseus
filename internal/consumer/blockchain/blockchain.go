// Package blockchain is a consumer interface.
package blockchain

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"

	"github.com/Decentr-net/ariadne"
	"github.com/Decentr-net/decentr/app"
	community "github.com/Decentr-net/decentr/x/community/types"
	"github.com/Decentr-net/decentr/x/operations"

	"github.com/Decentr-net/theseus/internal/consumer"
	"github.com/Decentr-net/theseus/internal/storage"
)

// nolint:gochecknoinits
func init() {
	c := sdk.GetConfig()
	c.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	c.Seal()
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

			for _, msg := range block.Messages() {
				var err error

				switch msg := msg.(type) {
				case community.MsgCreatePost:
					err = processMsgCreatePost(ctx, s, block.Time, &msg)
				case community.MsgDeletePost:
					err = processMsgDeletePost(ctx, s, block.Time, msg)
				case community.MsgSetLike:
					err = processMsgSetLike(ctx, s, block.Time, msg)
				case community.MsgFollow:
					err = processMsgFollow(ctx, s, msg)
				case community.MsgUnfollow:
					err = processMsgUnfollow(ctx, s, msg)
				case operations.MsgDistributeRewards:
					err = processDistributeRewards(ctx, s, block.Time, &msg)
				case operations.MsgResetAccount:
					err = processMsgResetAccount(ctx, s, msg.AccountOwner)
				default:
					log.WithField("msg", fmt.Sprintf("%s/%s", msg.Route(), msg.Type())).Debug("skip message")
				}

				if err != nil {
					return fmt.Errorf("failed to process msg: %w", err)
				}
			}

			if err := s.SetHeight(ctx, block.Height); err != nil {
				return fmt.Errorf("failed to set height: %w", err)
			}

			if err := s.RefreshViews(ctx); err != nil {
				return fmt.Errorf("failed to refresh views: %w", err)
			}

			return nil
		})
	}
}

func processMsgCreatePost(ctx context.Context, s storage.Storage, timestamp time.Time, msg *community.MsgCreatePost) error {
	return s.CreatePost(ctx, &storage.CreatePostParams{
		UUID:         msg.UUID,
		Owner:        msg.Owner.String(),
		Title:        msg.Title,
		Category:     msg.Category,
		PreviewImage: msg.PreviewImage,
		Text:         msg.Text,
		CreatedAt:    timestamp,
	})
}

func processMsgDeletePost(ctx context.Context, s storage.Storage, timestamp time.Time, msg community.MsgDeletePost) error {
	if err := s.DeletePost(ctx, storage.PostID{Owner: msg.PostOwner.String(), UUID: msg.PostUUID}, timestamp, msg.Owner.String()); err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return err
		}
	}
	return nil
}

func processMsgSetLike(ctx context.Context, s storage.Storage, timestamp time.Time, msg community.MsgSetLike) error {
	p := storage.PostID{
		Owner: msg.PostOwner.String(),
		UUID:  msg.PostUUID,
	}

	m, err := s.GetLikes(ctx, msg.Owner.String(), p)
	if err != nil {
		return fmt.Errorf("failed to get like: %w", err)
	}

	previousWeight := community.LikeWeightZero
	if l, ok := m[p]; ok {
		previousWeight = l
	}

	if err := s.AddPDV(ctx, msg.PostOwner.String(), int64(msg.Weight-previousWeight), timestamp); err != nil {
		return fmt.Errorf("failed to add pdv to profile stats: %w", err)
	}

	postID := storage.PostID{Owner: msg.PostOwner.String(), UUID: msg.PostUUID}

	// check the related post exists
	if _, err := s.GetPost(ctx, postID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.WithField("msg", msg).Errorf("set like: related post not found")
			return nil
		}
		return err
	}

	return s.SetLike(ctx, postID, msg.Weight, timestamp, msg.Owner.String())
}

func processMsgFollow(ctx context.Context, s storage.Storage, msg community.MsgFollow) error {
	return s.Follow(ctx, msg.Owner.String(), msg.Whom.String())
}

func processMsgUnfollow(ctx context.Context, s storage.Storage, msg community.MsgUnfollow) error {
	return s.Unfollow(ctx, msg.Owner.String(), msg.Whom.String())
}

func processDistributeRewards(ctx context.Context, s storage.Storage, timestamp time.Time, msg *operations.MsgDistributeRewards) error {
	for _, v := range msg.Rewards { // nolint:gocritic
		if err := s.AddPDV(ctx, v.Receiver.String(), int64(v.Reward), timestamp); err != nil {
			return fmt.Errorf("failed to add pdv: %w", err)
		}
	}

	return nil
}

func processMsgResetAccount(ctx context.Context, s storage.Storage, owner sdk.AccAddress) error {
	if err := s.ResetAccount(ctx, owner.String()); err != nil {
		return fmt.Errorf("failed to wipe account: %w", err)
	}

	return nil
}
