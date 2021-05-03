//+build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	m "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	community "github.com/Decentr-net/decentr/x/community/types"
	"github.com/Decentr-net/decentr/x/utils"

	"github.com/Decentr-net/theseus/internal/storage"
)

var (
	db  *sql.DB
	ctx = context.Background()
	s   storage.Storage
)

func TestMain(m *testing.M) {
	shutdown := setup()

	s = New(db)

	code := m.Run()
	shutdown()
	os.Exit(code)
}

func setup() func() {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:12",
		Env:          map[string]string{"POSTGRES_PASSWORD": "root"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
	})
	if err != nil {
		logrus.WithError(err).Fatalf("failed to create container")
	}

	if err := c.Start(ctx); err != nil {
		logrus.WithError(err).Fatal("failed to start container")
	}

	host, err := c.Host(ctx)
	if err != nil {
		logrus.WithError(err).Fatal("failed to get host")
	}

	port, err := c.MappedPort(ctx, "5432")
	if err != nil {
		logrus.WithError(err).Fatal("failed to map port")
	}

	dsn := fmt.Sprintf("host=%s port=%d user=postgres password=root sslmode=disable", host, port.Int())

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		logrus.WithError(err).Fatal("failed to open connection")
	}

	if err := db.Ping(); err != nil {
		logrus.WithError(err).Fatal("failed to ping postgres")
	}

	shutdownFn := func() {
		if c != nil {
			c.Terminate(ctx)
		}
	}

	migrate("postgres", "root", host, "postgres", port.Int())

	return shutdownFn
}

func migrate(username, password, hostname, dbname string, port int) {
	_, currFile, _, ok := runtime.Caller(0)
	if !ok {
		logrus.Fatal("failed to get current file location")
	}

	migrations := filepath.Join(currFile, "../../../../scripts/migrations/postgres/")

	migrator, err := m.New(
		fmt.Sprintf("file://%s", migrations),
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			username, password, hostname, port, dbname),
	)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create migrator")
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil {
		logrus.WithError(err).Fatal("failed to migrate")
	}
}

func cleanup(t *testing.T) {
	_, err := db.ExecContext(ctx, `UPDATE height SET height=0`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `DELETE FROM "like"`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `DELETE FROM post`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `DELETE FROM updv`)
	require.NoError(t, err)

	refreshViews(t)
}

func refreshViews(t *testing.T) {
	_, err := db.ExecContext(ctx, `REFRESH MATERIALIZED VIEW calculated_post`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `REFRESH MATERIALIZED VIEW stats`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `REFRESH MATERIALIZED VIEW pdv_stats`)
	require.NoError(t, err)
}

func TestPg_GetHeight(t *testing.T) {
	defer cleanup(t)

	h, err := s.GetHeight(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, 0, h)
}

func TestPg_OnLockedHeight_Errors(t *testing.T) {
	defer cleanup(t)
	require.True(t, errors.Is(s.WithLockedHeight(context.Background(), 0, func(locked storage.Storage) error { return nil }), storage.ErrRequestedHeightIsTooLow))
	require.True(t, errors.Is(s.WithLockedHeight(context.Background(), 2, func(locked storage.Storage) error { return nil }), storage.ErrRequestedHeightIsTooHigh))
}

func TestPg_WithLockedHeight(t *testing.T) {
	defer cleanup(t)

	mu := sync.Mutex{}

	// Lock mutex to be sure if routine is started
	mu.Lock()
	go require.NoError(t, s.WithLockedHeight(context.Background(), 1, func(locked storage.Storage) error {
		mu.Unlock()                        // allow main routine execution
		time.Sleep(time.Millisecond * 500) // next OnLockHeight or GetHeight should wait

		h, err := locked.GetHeight(context.Background())
		require.NoError(t, err)
		require.EqualValues(t, 0, h)

		return nil
	}))

	mu.Lock() // there we lock to prevent execution continuing

	go func() {
		mu.Lock()         // wait until second WithLockedHeight will start
		defer mu.Unlock() // allow test to finish

		h, err := s.GetHeight(context.Background())
		require.NoError(t, err)
		require.EqualValues(t, 2, h)
	}()

	require.NoError(t, s.WithLockedHeight(context.Background(), 2, func(locked storage.Storage) error {
		mu.Unlock()                        // allow second routine to start
		time.Sleep(time.Millisecond * 500) // to be sure that second routine is started and GetHeight is called

		h, err := locked.GetHeight(context.Background())
		require.NoError(t, err)
		require.EqualValues(t, 1, h)

		return nil
	}))

	mu.Lock() // do not finish until second routine will finish
}

func TestPg_GetProfileStats(t *testing.T) {
	defer cleanup(t)

	require.NoError(t, s.CreatePost(ctx, &storage.CreatePostParams{
		Owner: "address",
		UUID:  "123",
	}))
	require.NoError(t, s.CreatePost(ctx, &storage.CreatePostParams{
		Owner: "address",
		UUID:  "124",
	}))
	require.NoError(t, s.DeletePost(ctx, storage.PostID{"address", "124"}, time.Now(), "address_2"))

	now := time.Now()
	yersterday := time.Now().Add(-time.Hour * 24)

	require.NoError(t, s.AddPDV(ctx, "address", utils.InitialTokenBalance().Int64(), time.Time{}))
	require.NoError(t, s.AddPDV(ctx, "address_1", utils.InitialTokenBalance().Int64(), time.Time{}))

	require.NoError(t, s.AddPDV(ctx, "address", 10, yersterday))
	require.NoError(t, s.AddPDV(ctx, "address", 10, now))
	require.NoError(t, s.AddPDV(ctx, "address_1", 10, yersterday))

	refreshViews(t)

	pp, err := s.GetProfileStats(ctx, "address", "address_1", "address_2")
	require.NoError(t, err)
	require.Len(t, pp, 3)

	assert.EqualValues(t, &storage.ProfileStats{
		Address:    "address",
		PostsCount: 1,
		Stats: storage.Stats{
			"0001-01-01":                    1000000,
			yersterday.Format("2006-01-02"): 1000010,
			now.Format("2006-01-02"):        1000020,
		},
	}, pp[0])
	assert.EqualValues(t, &storage.ProfileStats{
		Address:    "address_1",
		PostsCount: 0,
		Stats: storage.Stats{
			"0001-01-01":                    1000000,
			yersterday.Format("2006-01-02"): 1000010,
		},
	}, pp[1])
	assert.EqualValues(t, &storage.ProfileStats{
		Address:    "address_2",
		PostsCount: 0,
		Stats:      storage.Stats{},
	}, pp[2])
}

func TestPg_CreatePost(t *testing.T) {
	defer cleanup(t)

	expected := storage.CreatePostParams{
		UUID:         "1",
		Owner:        "2",
		Title:        "3",
		Category:     4,
		PreviewImage: "5",
		Text:         "6",
		CreatedAt:    time.Now(),
	}

	require.NoError(t, s.CreatePost(ctx, &expected))
	refreshViews(t)

	p, err := s.GetPost(ctx, storage.PostID{expected.Owner, expected.UUID})
	require.NoError(t, err)
	require.Equal(t, expected.Owner, p.Owner)
	require.Equal(t, expected.UUID, p.UUID)
	require.Equal(t, expected.Title, p.Title)
	require.Equal(t, expected.Category, p.Category)
	require.Equal(t, expected.PreviewImage, p.PreviewImage)
	require.Equal(t, expected.Text, p.Text)
	require.Equal(t, expected.CreatedAt.UTC().Unix(), p.CreatedAt.Unix())
}

func TestPg_GetPost(t *testing.T) {
	defer cleanup(t)

	// GetPost tested in other tests

	_, err := s.GetPost(ctx, storage.PostID{"1", "2"})
	require.Equal(t, storage.ErrNotFound, err)
}

func TestPg_DeletePost(t *testing.T) {
	defer cleanup(t)

	p := storage.CreatePostParams{
		UUID:         "1",
		Owner:        "2",
		Title:        "3",
		Category:     4,
		PreviewImage: "5",
		Text:         "6",
		CreatedAt:    time.Now(),
	}

	require.NoError(t, s.CreatePost(ctx, &p))
	refreshViews(t)

	require.NoError(t, s.DeletePost(ctx, storage.PostID{p.Owner, p.UUID}, p.CreatedAt, "moderator"))
	refreshViews(t)

	_, err := s.GetPost(ctx, storage.PostID{p.Owner, p.UUID})
	require.Equal(t, storage.ErrNotFound, err)

	var info struct {
		DeletedAt time.Time `db:"deleted_at"`
		DeletedBy string    `db:"deleted_by"`
	}

	require.NoError(t, sqlx.Get(sqlx.NewDb(db, "postgres"), &info,
		`SELECT deleted_at, deleted_by FROM post WHERE owner=$1 AND uuid=$2`,
		p.Owner, p.UUID,
	))
	require.Equal(t, p.CreatedAt.UTC().Unix(), info.DeletedAt.Unix())
	require.Equal(t, "moderator", info.DeletedBy)
}

func TestPg_GetLiked(t *testing.T) {
	defer cleanup(t)

	require.NoError(t, s.CreatePost(ctx, &storage.CreatePostParams{UUID: "1", Owner: "1", Category: 1, CreatedAt: time.Now()}))
	require.NoError(t, s.CreatePost(ctx, &storage.CreatePostParams{UUID: "2", Owner: "2", Category: 2, CreatedAt: time.Now()}))

	require.NoError(t, s.SetLike(ctx, storage.PostID{"1", "1"}, -1, time.Now(), "3"))

	refreshViews(t)

	likes, err := s.GetLikes(ctx, "3", storage.PostID{"1", "1"}, storage.PostID{"2", "2"})
	require.NoError(t, err)
	require.Len(t, likes, 1)

	require.Equal(t, community.LikeWeightDown, likes[storage.PostID{"1", "1"}])
}

func TestPg_SetLike(t *testing.T) {
	defer cleanup(t)

	require.Equal(t, storage.ErrNotFound, s.SetLike(ctx, storage.PostID{"1", "2"}, 1, time.Now(), "liker"))

	p := storage.CreatePostParams{
		UUID:         "1",
		Owner:        "2",
		Title:        "3",
		Category:     4,
		PreviewImage: "5",
		Text:         "6",
		CreatedAt:    time.Now().UTC(),
	}

	require.NoError(t, s.CreatePost(ctx, &p))
	require.NoError(t, s.SetLike(ctx, storage.PostID{p.Owner, p.UUID}, 1, p.CreatedAt, "liker"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{p.Owner, p.UUID}, -1, p.CreatedAt, "liker2"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{p.Owner, p.UUID}, -1, p.CreatedAt, "liker3"))
	refreshViews(t)

	post, err := s.GetPost(ctx, storage.PostID{p.Owner, p.UUID})
	require.NoError(t, err)

	require.EqualValues(t, 1, post.Likes)
	require.EqualValues(t, 2, post.Dislikes)
	require.EqualValues(t, -1, post.UPDV)
}

func TestPg_Follow(t *testing.T) {
	defer cleanup(t)

	require.NoError(t, s.Follow(ctx, "1", "2"))

	var f struct {
		Follower string `db:"follower"`
		Followee string `db:"followee"`
	}

	require.NoError(t, sqlx.NewDb(db, "postgres").GetContext(ctx, &f, `SELECT * FROM follow`))

	require.Equal(t, "1", f.Follower)
	require.Equal(t, "2", f.Followee)
}

func TestPg_Unfollow(t *testing.T) {
	defer cleanup(t)

	require.NoError(t, s.Follow(ctx, "1", "2"))
	require.NoError(t, s.Unfollow(ctx, "1", "2"))

	var f struct {
		Follower string `db:"follower"`
		Followee string `db:"followee"`
	}

	err := sqlx.NewDb(db, "postgres").GetContext(ctx, &f, `SELECT * FROM follow`)
	require.Error(t, err)
	require.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestPg_ListPosts(t *testing.T) {
	defer cleanup(t)

	require.NoError(t, s.CreatePost(ctx, &storage.CreatePostParams{UUID: "1", Owner: "1", Category: 1, CreatedAt: time.Unix(1, 0)}))
	require.NoError(t, s.CreatePost(ctx, &storage.CreatePostParams{UUID: "2", Owner: "2", Category: 2, CreatedAt: time.Unix(2, 0)}))
	require.NoError(t, s.CreatePost(ctx, &storage.CreatePostParams{UUID: "3", Owner: "3", Category: 3, CreatedAt: time.Unix(3, 0)}))
	require.NoError(t, s.CreatePost(ctx, &storage.CreatePostParams{UUID: "4", Owner: "4", Category: 4, CreatedAt: time.Unix(4, 0)}))
	require.NoError(t, s.CreatePost(ctx, &storage.CreatePostParams{UUID: "5", Owner: "5", Category: 5, CreatedAt: time.Unix(5, 0)}))

	require.NoError(t, s.Follow(ctx, "1", "2"))
	require.NoError(t, s.Follow(ctx, "1", "3"))

	require.NoError(t, s.SetLike(ctx, storage.PostID{"5", "5"}, 1, time.Unix(1, 0), "13"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"5", "5"}, 1, time.Unix(1, 0), "3"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"5", "5"}, -1, time.Unix(1, 0), "4"))

	require.NoError(t, s.SetLike(ctx, storage.PostID{"1", "1"}, 1, time.Unix(1, 0), "3"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"1", "1"}, 1, time.Unix(1, 0), "4"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"1", "1"}, 1, time.Unix(1, 0), "13"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"1", "1"}, -1, time.Unix(1, 0), "51"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"1", "1"}, -1, time.Unix(1, 0), "5"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"1", "1"}, -1, time.Unix(1, 0), "6"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"1", "1"}, -1, time.Unix(1, 0), "7"))

	require.NoError(t, s.SetLike(ctx, storage.PostID{"2", "2"}, 1, time.Unix(1, 0), "2"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"2", "2"}, 1, time.Unix(1, 0), "22"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"2", "2"}, 1, time.Unix(1, 0), "3"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"2", "2"}, 1, time.Unix(1, 0), "4"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"2", "2"}, -1, time.Unix(1, 0), "12"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"2", "2"}, -1, time.Unix(1, 0), "13"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"2", "2"}, -1, time.Unix(1, 0), "14"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"2", "2"}, -1, time.Unix(1, 0), "15"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"2", "2"}, -1, time.Unix(1, 0), "16"))

	require.NoError(t, s.SetLike(ctx, storage.PostID{"4", "4"}, 1, time.Unix(1, 0), "2"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"4", "4"}, 1, time.Unix(1, 0), "21"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"4", "4"}, 1, time.Unix(1, 0), "3"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"4", "4"}, 1, time.Unix(1, 0), "4"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"4", "4"}, 1, time.Unix(1, 0), "5"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"4", "4"}, -1, time.Unix(1, 0), "12"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"4", "4"}, -1, time.Unix(1, 0), "13"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"4", "4"}, -1, time.Unix(1, 0), "14"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"4", "4"}, -1, time.Unix(1, 0), "15"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"4", "4"}, -1, time.Unix(1, 0), "16"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"4", "4"}, -1, time.Unix(1, 0), "17"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"4", "4"}, -1, time.Unix(1, 0), "18"))

	refreshViews(t)

	cat := community.Category(3)
	owner := "2"
	likedBy := "5"
	followedBy := "1"
	from := uint64(2)
	to := uint64(5)
	after := storage.PostID{Owner: "2", UUID: "2"}

	tt := []struct {
		name string
		p    storage.ListPostsParams
		ids  []string
	}{
		{
			name: "created_at_asc",
			p: storage.ListPostsParams{
				SortBy:  storage.CreatedAtSortType,
				OrderBy: storage.AscendingOrder,
				Limit:   100,
			},
			ids: []string{"1", "2", "3", "4", "5"},
		},
		{
			name: "likes_desc",
			p: storage.ListPostsParams{
				SortBy:  storage.LikesSortType,
				OrderBy: storage.DescendingOrder,
				Limit:   100,
			},
			ids: []string{"4", "2", "1", "5", "3"},
		},
		{
			name: "dislikes_desc",
			p: storage.ListPostsParams{
				SortBy:  storage.DislikesSortType,
				OrderBy: storage.DescendingOrder,
				Limit:   100,
			},
			ids: []string{"4", "2", "1", "5", "3"},
		},
		{
			name: "pdv_desc",
			p: storage.ListPostsParams{
				SortBy:  storage.PDVSortType,
				OrderBy: storage.DescendingOrder,
				Limit:   100,
			},
			ids: []string{"5", "3", "2", "1", "4"},
		},
		{
			name: "category",
			p: storage.ListPostsParams{
				SortBy:   storage.CreatedAtSortType,
				OrderBy:  storage.DescendingOrder,
				Category: &cat,
				Limit:    100,
			},
			ids: []string{"3"},
		},
		{
			name: "owner",
			p: storage.ListPostsParams{
				SortBy:  storage.LikesSortType,
				OrderBy: storage.DescendingOrder,
				Limit:   100,
				Owner:   &owner,
			},
			ids: []string{"2"},
		},
		{
			name: "liked_by",
			p: storage.ListPostsParams{
				SortBy:  storage.LikesSortType,
				OrderBy: storage.DescendingOrder,
				Limit:   100,
				LikedBy: &likedBy,
			},
			ids: []string{"4", "1"},
		},
		{
			name: "followed_by",
			p: storage.ListPostsParams{
				SortBy:     storage.CreatedAtSortType,
				OrderBy:    storage.AscendingOrder,
				Limit:      100,
				FollowedBy: &followedBy,
			},
			ids: []string{"2", "3"},
		},
		{
			name: "from_to",
			p: storage.ListPostsParams{
				SortBy:  storage.CreatedAtSortType,
				OrderBy: storage.AscendingOrder,
				Limit:   100,
				From:    &from,
				To:      &to,
			},
			ids: []string{"3", "4"},
		},
		{
			name: "after",
			p: storage.ListPostsParams{
				SortBy:  storage.CreatedAtSortType,
				OrderBy: storage.AscendingOrder,
				Limit:   100,
				After:   &after,
			},
			ids: []string{"3", "4", "5"},
		},
		{
			name: "after_same_value_desc",
			p: storage.ListPostsParams{
				SortBy:  storage.PDVSortType,
				OrderBy: storage.DescendingOrder,
				Limit:   100,
				After:   &after,
			},
			ids: []string{"1", "4"},
		},
		{
			name: "after_same_value_asc",
			p: storage.ListPostsParams{
				SortBy:  storage.PDVSortType,
				OrderBy: storage.AscendingOrder,
				Limit:   100,
				After:   &after,
			},
			ids: []string{"3", "5"},
		},
	}

	for i := range tt {
		tc := tt[i]
		t.Run(tc.name, func(t *testing.T) {
			p, err := s.ListPosts(ctx, &tc.p)
			require.NoError(t, err)
			require.Len(t, p, len(tc.ids))
			for i, v := range tc.ids {
				require.Equal(t, v, p[i].UUID)
			}
		})
	}
}

func TestPg_GetStats(t *testing.T) {
	defer cleanup(t)

	today := time.Now().UTC()
	yesterday := today.Add(-time.Hour * 24)
	monthAgo := today.Add(-time.Hour * 24 * 32)

	require.NoError(t, s.CreatePost(ctx, &storage.CreatePostParams{UUID: "1", Owner: "1", Category: 1, CreatedAt: time.Unix(1, 0)}))
	require.NoError(t, s.CreatePost(ctx, &storage.CreatePostParams{UUID: "2", Owner: "2", Category: 2, CreatedAt: time.Unix(2, 0)}))

	require.NoError(t, s.SetLike(ctx, storage.PostID{"1", "1"}, 1, today, "3"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"1", "1"}, 1, monthAgo, "4"))

	require.NoError(t, s.SetLike(ctx, storage.PostID{"2", "2"}, 1, today, "2"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"2", "2"}, 1, today, "5"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"2", "2"}, 1, yesterday, "3"))
	require.NoError(t, s.SetLike(ctx, storage.PostID{"2", "2"}, 1, monthAgo, "4"))

	refreshViews(t)

	stats, err := s.GetPostStats(ctx, storage.PostID{"1", "1"}, storage.PostID{"2", "2"})
	require.NoError(t, err)

	// nolint
	assert.Equal(t, map[storage.PostID]storage.Stats{
		storage.PostID{"1", "1"}: {
			today.Format("2006-01-02"):    2,
			monthAgo.Format("2006-01-02"): 1,
		},
		storage.PostID{"2", "2"}: {
			today.Format("2006-01-02"):     4,
			yesterday.Format("2006-01-02"): 2,
			monthAgo.Format("2006-01-02"):  1,
		},
	}, stats)
}

func TestPg_AddPDV(t *testing.T) {
	defer cleanup(t)

	require.NoError(t, s.AddPDV(ctx, "addr", 10, time.Now()))
}

func TestPg_GetAllUsersStats(t *testing.T) {
	defer cleanup(t)

	today := time.Now().UTC()
	yesterday := today.Add(-time.Hour * 24)
	monthAgo := today.Add(-time.Hour * 24 * 32)
	require.NoError(t, s.AddPDV(ctx, "addr", utils.InitialTokenBalance().Int64(), time.Time{}))
	require.NoError(t, s.AddPDV(ctx, "addr2", utils.InitialTokenBalance().Int64(), time.Time{}))
	require.NoError(t, s.AddPDV(ctx, "addr2", 5, today))
	require.NoError(t, s.AddPDV(ctx, "addr2", -15, yesterday))
	require.NoError(t, s.AddPDV(ctx, "addr", 10, today))
	require.NoError(t, s.AddPDV(ctx, "addr", 10, yesterday))
	require.NoError(t, s.AddPDV(ctx, "addr", 10, monthAgo))

	stats, err := s.GetDecentrStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, &storage.DecentrStats{
		ADV: 1000010,
		DDV: 20,
	}, stats)
}
