package blockchain

import (
	"context"
	"fmt"

	"github.com/Decentr-net/ariadne"
	"github.com/Decentr-net/decentr/x/community"
	"github.com/sirupsen/logrus"

	"github.com/Decentr-net/theseus/internal/consumer"
	"github.com/Decentr-net/theseus/internal/service"
	"github.com/Decentr-net/theseus/internal/storage"
)

type HeightLocker interface {
	OnHeight(height uint64, f func(s storage.Storage) error) error
}

var log = logrus.WithField("package", "blockchain")

type blockchain struct {
	f ariadne.Fetcher
	l HeightLocker
}

func New(f ariadne.Fetcher, l HeightLocker) consumer.Consumer {
	return blockchain{
		f: f,
		l: l,
	}
}

func logError(h uint64, err error) {
	log.WithField("height", h).WithError(err).Error("failed to process block")
}

func (b blockchain) Run(ctx context.Context, from uint64) error {
	ch := b.f.FetchBlocks(ctx, from, ariadne.WithErrHandler(logError))

	for {
		select {
		case <-ctx.Done():
			return nil
		case block := <-ch:
			if err := b.processBlock(ctx, block); err != nil {
				logError(block.Height, err)
			}
		}
	}
}

func (b blockchain) processBlock(ctx context.Context, block ariadne.Block) error {
	err := b.l.OnHeight(block.Height, func(s storage.Storage) error {
		srv := service.New(s)
		for _, msg := range block.Messages() {
			switch msg := msg.(type) {
			case community.MsgCreatePost:
				if err := srv.CreatePost(ctx, service.Post{
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
				if err := srv.DeletePost(ctx, msg.PostOwner.String(), msg.PostUUID, block.Time, msg.Owner.String()); err != nil {
					return err
				}

			case community.MsgSetLike:
				if err := srv.SetLike(ctx, msg.PostOwner.String(), msg.PostUUID, msg.Weight, block.Time, msg.Owner.String()); err != nil {
					return err
				}

			default:
				log.WithField("msg", fmt.Sprintf("%s/%s", msg.Route(), msg.Type())).Debug("skip message")
			}
		}

		return nil
	})

	return err
}
