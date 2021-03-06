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
	"github.com/Decentr-net/decentr/x/pdv"
	"github.com/Decentr-net/decentr/x/profile"
	"github.com/Decentr-net/decentr/x/utils"

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
		err := b.s.WithLockedHeight(ctx, block.Height, func(s storage.Storage) error {
			log.WithField("height", block.Height).WithField("txs", len(block.Messages())).Info("processing block")

			for _, msg := range block.Messages() {
				switch msg := msg.(type) {
				case profile.MsgSetPublic:
					return processMsgSetPublicProfile(ctx, s, block.Time, &msg)
				case community.MsgCreatePost:
					return processMsgCreatePost(ctx, s, block.Time, &msg)
				case community.MsgDeletePost:
					return processMsgDeletePost(ctx, s, block.Time, msg)
				case community.MsgSetLike:
					return processMsgSetLike(ctx, s, block.Time, msg)
				case community.MsgFollow:
					return processMsgFollow(ctx, s, msg)
				case community.MsgUnfollow:
					return processMsgUnfollow(ctx, s, msg)
				case pdv.MsgDistributeRewards:
					return processDistributeRewards(ctx, s, block.Time, &msg)
				default:
					log.WithField("msg", fmt.Sprintf("%s/%s", msg.Route(), msg.Type())).Debug("skip message")
				}
			}

			return nil
		})

		// A block is processed, we shouldn't retry processing
		if errors.Is(err, storage.ErrRequestedHeightIsTooLow) {
			return nil
		}

		return err
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
	return s.DeletePost(ctx, storage.PostID{Owner: msg.PostOwner.String(), UUID: msg.PostUUID}, timestamp, msg.Owner.String())
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

	return s.SetLike(ctx, storage.PostID{Owner: msg.PostOwner.String(), UUID: msg.PostUUID}, msg.Weight, timestamp, msg.Owner.String())
}

func processMsgSetPublicProfile(ctx context.Context, s storage.Storage, timestamp time.Time, msg *profile.MsgSetPublic) error {
	if _, err := s.GetProfileStats(ctx, msg.Owner.String()); err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("failed to check profile's pdv: %w", err)
		}
		if err := s.AddPDV(ctx, msg.Owner.String(), utils.InitialTokenBalance().Int64(), timestamp); err != nil {
			return fmt.Errorf("failed to set initial pdv balance: %w", err)
		}
	}

	if err := s.SetProfile(ctx, &storage.SetProfileParams{
		Address:   msg.Owner.String(),
		FirstName: msg.Public.FirstName,
		LastName:  msg.Public.LastName,
		Bio:       msg.Public.Bio,
		Avatar:    msg.Public.Avatar,
		Gender:    string(msg.Public.Gender),
		Birthday:  msg.Public.Birthday,
		CreatedAt: timestamp,
	}); err != nil {
		return fmt.Errorf("failed to set profile: %w", err)
	}

	return nil
}

func processMsgFollow(ctx context.Context, s storage.Storage, msg community.MsgFollow) error {
	return s.Follow(ctx, msg.Owner.String(), msg.Whom.String())
}

func processMsgUnfollow(ctx context.Context, s storage.Storage, msg community.MsgUnfollow) error {
	return s.Unfollow(ctx, msg.Owner.String(), msg.Whom.String())
}

func processDistributeRewards(ctx context.Context, s storage.Storage, timestamp time.Time, msg *pdv.MsgDistributeRewards) error {
	for _, v := range msg.Rewards { // nolint:gocritic
		if err := s.AddPDV(ctx, v.Receiver.String(), int64(v.Reward), timestamp); err != nil {
			return fmt.Errorf("failed to add pdv: %w", err)
		}
	}

	return nil
}
