// Package postgres is implementation of storage interface.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"

	community "github.com/Decentr-net/decentr/x/community/types"

	"github.com/Decentr-net/theseus/internal/entities"
	"github.com/Decentr-net/theseus/internal/storage"
)

var log = logrus.WithField("layer", "storage").WithField("package", "postgres")
var errBeginCalledWithinTx = errors.New("can not run WithLockedHeight in tx")

const foreignKeyViolation = "23503"

type pg struct {
	ext sqlx.ExtContext
}

type postDTO struct {
	UUID         string    `db:"uuid"`
	Owner        string    `db:"owner"`
	Title        string    `db:"title"`
	Category     uint8     `db:"category"`
	PreviewImage string    `db:"preview_image"`
	Text         string    `db:"text"`
	CreatedAt    time.Time `db:"created_at"`
}

func (s pg) WithLockedHeight(ctx context.Context, height uint64, f func(s storage.Storage) error) error {
	db, ok := s.ext.(*sqlx.DB)
	if !ok {
		return errBeginCalledWithinTx
	}

	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("failed to create tx: %w", err)
	}

	if err := func(s storage.Storage) error {
		// WithLockedHeight should be blocking method
		if _, err := tx.ExecContext(ctx, `LOCK TABLE height IN ACCESS EXCLUSIVE MODE`); err != nil {
			return fmt.Errorf("failed to lock height table: %w", err)
		}

		h, err := s.GetHeight(ctx)
		if err != nil {
			return fmt.Errorf("failed to get height: %w", err)
		}

		if height > h+1 {
			return fmt.Errorf("%w expected_height=%d", storage.ErrRequestedHeightIsTooHigh, h+1)
		}

		if height < h+1 {
			return fmt.Errorf("%w expected_height=%d", storage.ErrRequestedHeightIsTooLow, h+1)
		}

		if err := f(s); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `UPDATE height SET height=$1`, height); err != nil {
			return fmt.Errorf("failed to save height: failed to exec: %w", err)
		}

		return nil
	}(pg{ext: tx}); err != nil {
		if err := tx.Rollback(); err != nil {
			log.WithError(err).Error("failed to rollback tx")
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commint tx: %w", err)
	}

	return nil
}

func (s pg) GetHeight(ctx context.Context) (uint64, error) {
	var h uint64
	if err := sqlx.GetContext(ctx, s.ext, &h, `SELECT height FROM height FOR KEY SHARE`); err != nil {
		return 0, fmt.Errorf("failed to query: %w", err)
	}

	return h, nil
}

func (s pg) SetHeight(ctx context.Context, h uint64) error {
	if _, err := s.ext.ExecContext(ctx, `UPDATE height SET height=$1`, h); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

func (s pg) CreatePost(ctx context.Context, p *entities.Post) error {
	post := postDTO{
		UUID:         p.UUID,
		Owner:        p.Owner,
		Title:        p.Title,
		Category:     uint8(p.Category),
		PreviewImage: p.PreviewImage,
		Text:         p.Text,
		CreatedAt:    p.CreatedAt.UTC(),
	}

	if _, err := sqlx.NamedExecContext(ctx, s.ext,
		`
			INSERT INTO post(owner, uuid, title, category, preview_image, text, created_at)
			VALUES(:owner, :uuid, :title, :category, :preview_image, :text, :created_at)
		`, post,
	); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

func (s pg) GetPost(ctx context.Context, owner, uuid string) (*entities.Post, error) {
	var p postDTO

	if err := sqlx.GetContext(ctx, s.ext, &p, `
			SELECT owner, uuid, title, category, preview_image, text, created_at
			FROM post
			WHERE owner = $1 AND uuid = $2 AND deleted_at IS NULL
		`,
		owner, uuid,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNotFound
		}

		return nil, fmt.Errorf("failed to query: %w", err)
	}

	return &entities.Post{
		UUID:         p.UUID,
		Owner:        p.Owner,
		Title:        p.Title,
		Category:     community.Category(p.Category),
		PreviewImage: p.PreviewImage,
		Text:         p.Text,
		CreatedAt:    p.CreatedAt,
	}, nil
}

func (s pg) DeletePost(ctx context.Context, postOwner string, postUUID string, timestamp time.Time, deletedBy string) error {
	res, err := s.ext.ExecContext(ctx,
		`UPDATE post SET deleted_at=$3, deleted_by=$4 WHERE owner=$1 AND uuid=$2 AND deleted_at IS NULL`,
		postOwner, postUUID, timestamp.UTC(), deletedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	if c, _ := res.RowsAffected(); c == 0 {
		return storage.ErrNotFound
	}

	return nil
}

func (s pg) SetLike(ctx context.Context, postOwner string, postUUID string, weight community.LikeWeight,
	timestamp time.Time, likeOwner string) error {
	if _, err := s.ext.ExecContext(ctx, `
			INSERT INTO "like"(post_owner, post_uuid, liked_by, weight, liked_at)
				VALUES($1, $2, $3, $4, $5)
			ON CONFLICT(post_owner, post_uuid, liked_by) DO UPDATE SET
				weight=excluded.weight, liked_at=excluded.liked_at`,
		postOwner, postUUID, likeOwner, weight, timestamp.UTC(),
	); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == foreignKeyViolation {
			return storage.ErrNotFound
		}

		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

// New creates new instance of pg.
func New(db *sql.DB) storage.Storage {
	return pg{
		ext: sqlx.NewDb(db, "postgres"),
	}
}
