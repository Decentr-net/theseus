package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/Decentr-net/decentr/app"
	"github.com/Decentr-net/decentr/x/community"
	"github.com/Decentr-net/decentr/x/profile"
	"github.com/Decentr-net/decentr/x/token"
	"github.com/golang-migrate/migrate/v4"
	migratep "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jessevdk/go-flags"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	_ "github.com/Decentr-net/theseus/internal/consumer/blockchain"
	"github.com/Decentr-net/theseus/internal/storage"
	"github.com/Decentr-net/theseus/internal/storage/postgres"
)

var opts = struct {
	Genesis            string `long:"genesis" env:"GENESIS" default:"genesis.json" description:"path to genesis"`
	Postgres           string `long:"postgres" env:"POSTGRES" default:"host=localhost port=5432 user=postgres password=root sslmode=disable" description:"postgres dsn"`
	PostgresMigrations string `long:"postgres.migrations" env:"POSTGRES_MIGRATIONS" default:"scripts/migrations/postgres" description:"postgres migrations directory"`
}{}

type genesis struct {
	AppState struct {
		Community community.GenesisState `json:"community"`
		Profile   profile.GenesisState   `json:"profile"`
		Token     token.GenesisState     `json:"token"`
	} `json:"app_state"`
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	parser.ShortDescription = "genesis2db"
	parser.LongDescription = "Genesis to database importer"

	_, err := parser.Parse()

	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			parser.WriteHelp(os.Stdout)
			os.Exit(0)
		}
		logrus.WithError(err).Fatal("error occurred while parsing flags")
	}

	logrus.Info("db2migration started")
	logrus.Infof("%+v", opts)

	b, err := ioutil.ReadFile(opts.Genesis)
	if err != nil {
		logrus.WithError(err).Fatal("failed to read genesis")
	}

	var g genesis

	cdc := app.MakeCodec()
	cdc.MustUnmarshalJSON(b, &g)

	db := mustGetDB()
	s := postgres.New(db)

	logrus.Info("import profiles")

	t := time.Now().UTC()

	for i, v := range g.AppState.Profile.ProfileRecords {
		if err := s.SetProfile(context.Background(), &storage.SetProfileParams{
			Address:   v.Owner.String(),
			FirstName: v.Public.FirstName,
			LastName:  v.Public.LastName,
			Bio:       v.Public.Bio,
			Avatar:    v.Public.Avatar,
			Gender:    string(v.Public.Gender),
			Birthday:  v.Public.Birthday,
			CreatedAt: t,
		}); err != nil {
			logrus.WithError(err).Fatal("failed to put profile into db")
		}

		if i%20 == 0 {
			logrus.Infof("%d of %d profiles imported", i+1, len(g.AppState.Profile.ProfileRecords))
		}
	}

	logrus.Info("import token")
	i := 0
	for k, v := range g.AppState.Token.Balances {
		if err := s.AddPDV(context.Background(), k, v.Int64(), t); err != nil {
			logrus.WithError(err).Fatal("failed to put token into db")
		}

		i++
		if i%20 == 0 {
			logrus.Infof("%d of %d balances imported", i+1, len(g.AppState.Token.Balances))
		}
	}

	i = 0
	logrus.Info("import followings")
	for follower, v := range g.AppState.Community.Followers {
		for _, followee := range v {
			if err := s.Follow(context.Background(), follower, followee); err != nil {
				logrus.WithError(err).Fatal("failed to put following into db")
			}
		}

		i++
		if i%20 == 0 {
			logrus.Infof("%d of %d followers imported", i+1, len(g.AppState.Community.Followers))
		}
	}

	logrus.Info("import posts")
	for i, v := range g.AppState.Community.Posts {
		if err := s.CreatePost(context.Background(), &storage.CreatePostParams{
			UUID:         v.UUID.String(),
			Owner:        v.Owner.String(),
			Title:        v.Title,
			Category:     v.Category,
			PreviewImage: v.PreviewImage,
			Text:         v.Text,
			CreatedAt:    t,
		}); err != nil {
			logrus.WithError(err).Fatal("failed to put post into db")
		}

		if i%20 == 0 {
			logrus.Infof("%d of %d posts imported", i+1, len(g.AppState.Community.Posts))
		}
	}

	logrus.Info("import likes")
	for i, v := range g.AppState.Community.Likes {
		if err := s.SetLike(context.Background(), storage.PostID{
			Owner: v.PostOwner.String(),
			UUID:  v.PostUUID.String(),
		}, v.Weight, t, v.Owner.String()); err != nil {
			logrus.WithError(err).Fatal("failed to put like into db")
		}

		if i%20 == 0 {
			logrus.Infof("%d of %d likes imported", i+1, len(g.AppState.Community.Likes))
		}
	}

	logrus.Info("refreshing posts view")
	if _, err := db.Exec(`REFRESH MATERIALIZED VIEW calculated_post`); err != nil {
		logrus.WithError(err).Fatal("failed to refresh posts view")
	}

	logrus.Info("done")
}

func mustGetDB() *sql.DB {
	db, err := sql.Open("postgres", opts.Postgres)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create postgres connection")
	}

	if err := db.PingContext(context.Background()); err != nil {
		logrus.WithError(err).Fatal("failed to ping postgres")
	}

	driver, err := migratep.WithInstance(db, &migratep.Config{})
	if err != nil {
		logrus.WithError(err).Fatal("failed to create database migrate driver")
	}

	migrator, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", opts.PostgresMigrations), "postgres", driver)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create migrator")
	}

	switch v, d, err := migrator.Version(); err {
	case nil:
		logrus.Infof("database version %d with dirty state %t", v, d)
	case migrate.ErrNilVersion:
		logrus.Info("database version: nil")
	default:
		logrus.WithError(err).Fatal("failed to get version")
	}

	switch err := migrator.Up(); err {
	case nil:
		logrus.Info("database was migrated")
	case migrate.ErrNoChange:
		logrus.Info("database is up-to-date")
	default:
		logrus.WithError(err).Fatal("failed to migrate db")
	}

	return db
}
