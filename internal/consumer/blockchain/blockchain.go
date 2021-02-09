// Package blockchain is a consumer interface.
package blockchain

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"

	"github.com/Decentr-net/ariadne"
	"github.com/Decentr-net/decentr/app"
	"github.com/Decentr-net/decentr/x/community"

	"github.com/Decentr-net/theseus/internal/consumer"
	"github.com/Decentr-net/theseus/internal/entities"
	"github.com/Decentr-net/theseus/internal/service"
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
	s service.Service

	retryInterval          time.Duration
	retryLastBlockInterval time.Duration
}

// New returns new blockchain instance.
func New(f ariadne.Fetcher, s service.Service) consumer.Consumer {
	return blockchain{
		f: f,
		s: s,
	}
}

func logError(h uint64, err error) {
	log.WithField("height", h).WithError(err).Error("failed to process block")
}

func (b blockchain) Run(ctx context.Context) error {
	from, err := b.s.GetHeight()
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
		return b.s.OnHeight(block.Height, func(s service.Service) error {
			for _, msg := range block.Messages() {
				switch msg := msg.(type) {
				case community.MsgCreatePost:

					if err := s.CreatePost(ctx, &entities.Post{
						UUID:         msg.UUID,
						Owner:        msg.Owner.String(),
						Title:        msg.Title,
						Category:     msg.Category,
						PreviewImage: msg.PreviewImage,
						Text:         msg.Text,
						CreatedAt:    block.Time,
					}); err != nil {
						return err
					}

				case community.MsgDeletePost:
					if err := s.DeletePost(ctx, msg.PostOwner.String(), msg.PostUUID, block.Time, msg.Owner.String()); err != nil {
						return err
					}

				case community.MsgSetLike:
					if err := s.SetLike(ctx, msg.PostOwner.String(), msg.PostUUID, msg.Weight, block.Time, msg.Owner.String()); err != nil {
						return err
					}

				default:
					log.WithField("msg", fmt.Sprintf("%s/%s", msg.Route(), msg.Type())).Debug("skip message")
				}
			}

			return nil
		})
	}
}
