// Package impl is implementation of service interface.
package impl

import (
	"context"
	"fmt"
	"time"

	"github.com/Decentr-net/decentr/x/community/types"

	"github.com/Decentr-net/theseus/internal/entities"
	"github.com/Decentr-net/theseus/internal/service"
	"github.com/Decentr-net/theseus/internal/storage"
)

// service ...
type srv struct {
	s storage.Storage
}

func (s srv) OnHeight(ctx context.Context, height uint64, f func(s service.Service) error) error {
	if err := s.s.OnLockedHeight(ctx, func(s storage.Storage) error {
		h, err := s.GetHeight(ctx)
		if err != nil {
			return fmt.Errorf("failed to get height on s side: %w", err)
		}

		if height > h+1 {
			return fmt.Errorf("%w height=%d", service.ErrRequestedHeightIsTooHigh, h)
		}

		if height != h+1 {
			return fmt.Errorf("%w height=%d", service.ErrRequestedHeightIsTooLow, h)
		}

		return f(New(s))
	}); err != nil {
		return err
	}

	return nil
}

func (s srv) GetHeight(ctx context.Context) (uint64, error) {
	h, err := s.s.GetHeight(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get height from s: %w", err)
	}

	return h, nil
}

func (s srv) CreatePost(ctx context.Context, p *entities.Post) error {
	if err := s.s.CreatePost(ctx, p); err != nil {
		return fmt.Errorf("failed to create post on s side: %w", err)
	}

	return nil
}

func (s srv) DeletePost(ctx context.Context, postOwner, postUUID string, timestamp time.Time, deletedByUUID string) error {
	if err := s.s.DeletePost(ctx, postOwner, postUUID, timestamp, deletedByUUID); err != nil {
		return fmt.Errorf("failed to delete post on s side: %w", err)
	}

	return nil
}

func (s srv) SetLike(ctx context.Context, postOwner, postUUID string, weight types.LikeWeight, timestamp time.Time, likedBy string) error {
	if err := s.s.SetLike(ctx, postOwner, postUUID, weight, timestamp, likedBy); err != nil {
		return fmt.Errorf("failed to set like on s side: %w", err)
	}

	return nil
}

// New creates new instance of service.
func New(s storage.Storage) service.Service {
	return srv{
		s: s,
	}
}
