package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang-migrate/migrate/v4"
	migratep "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jessevdk/go-flags"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/Decentr-net/ariadne"
	"github.com/Decentr-net/go-api/health"
	"github.com/Decentr-net/logrus/sentry"

	"github.com/Decentr-net/theseus/internal/consumer"
	"github.com/Decentr-net/theseus/internal/consumer/blockchain"
	"github.com/Decentr-net/theseus/internal/storage"
	"github.com/Decentr-net/theseus/internal/storage/postgres"
)

// nolint:lll,gochecknoglobals
var opts = struct {
	Host string `long:"http.host" env:"HTTP_HOST" default:"0.0.0.0" description:"IP to listen on"`
	Port int    `long:"http.port" env:"HTTP_PORT" default:"8080" description:"port to listen on for insecure connections, defaults to a random value"`

	Postgres                   string `long:"postgres" env:"POSTGRES" default:"host=localhost port=5432 user=postgres password=root sslmode=disable" description:"postgres dsn"`
	PostgresMaxOpenConnections int    `long:"postgres.max_open_connections" env:"POSTGRES_MAX_OPEN_CONNECTIONS" default:"0" description:"postgres maximal open connections count, 0 means unlimited"`
	PostgresMaxIdleConnections int    `long:"postgres.max_idle_connections" env:"POSTGRES_MAX_IDLE_CONNECTIONS" default:"5" description:"postgres maximal idle connections count"`
	PostgresMigrations         string `long:"postgres.migrations" env:"POSTGRES_MIGRATIONS" default:"migrations/postgres" description:"postgres migrations directory"`

	BlockchainNode                   string        `long:"blockchain.node" env:"BLOCKCHAIN_NODE" default:"zeus.testnet.decentr.xyz:9090" description:"decentr node address"`
	BlockchainTimeout                time.Duration `long:"blockchain.timeout" env:"BLOCKCHAIN_TIMEOUT" default:"5s" description:"timeout for requests to blockchain node"`
	BlockchainRetryInterval          time.Duration `long:"blockchain.retry_interval" env:"BLOCKCHAIN_RETRY_INTERVAL" default:"2s" description:"interval to be waited on error before retry"`
	BlockchainLastBlockRetryInterval time.Duration `long:"blockchain.last_block_retry_interval" env:"BLOCKCHAIN_LAST_BLOCK_RETRY_INTERVAL" default:"1s" description:"duration to be waited when new block isn't produced before retry"`

	LogLevel  string `long:"log.level" env:"LOG_LEVEL" default:"info" description:"Log level" choice:"debug" choice:"info" choice:"warning" choice:"error"`
	SentryDSN string `long:"sentry.dsn" env:"SENTRY_DSN" description:"sentry dsn"`
}{}

var errTerminated = errors.New("terminated")

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	parser.ShortDescription = "Theseus Sync"
	parser.LongDescription = "Theseus Sync"

	_, err := parser.Parse()

	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			parser.WriteHelp(os.Stdout)
			os.Exit(0)
		}
		logrus.WithError(err).Fatal("error occurred while parsing flags")
	}

	lvl, _ := logrus.ParseLevel(opts.LogLevel) // err will always be nil
	logrus.SetLevel(lvl)

	logrus.Info("service started")
	logrus.Infof("%+v", opts)

	if opts.SentryDSN != "" {
		hook, err := sentry.NewHook(sentry.Options{
			Dsn:              opts.SentryDSN,
			AttachStacktrace: true,
			Release:          health.GetVersion(),
			ServerName:       "sync",
		}, logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel)

		if err != nil {
			logrus.WithError(err).Fatal("failed to init sentry")
		}

		logrus.AddHook(hook)
	} else {
		logrus.Info("empty sentry dsn")
		logrus.Warn("skip sentry initialization")
	}

	db := mustGetDB()

	s := postgres.New(db)
	c := mustGetConsumer(s)

	r := chi.NewMux()
	r.Get("/health", health.Handler(
		5*time.Second,
		c, // consumer gets the height from db
	))
	srv := http.Server{
		Addr:    fmt.Sprintf("%s:%d", opts.Host, opts.Port),
		Handler: r,
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {}()

	gr, _ := errgroup.WithContext(ctx)
	gr.Go(func() error {
		return c.Run(ctx)
	})
	gr.Go(srv.ListenAndServe)
	gr.Go(func() error {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		s := <-sigs

		logrus.Infof("terminating by %s signal", s)

		cancel()

		return errTerminated
	})

	logrus.Info("service started")

	if err := gr.Wait(); err != nil && !errors.Is(err, errTerminated) {
		logrus.WithError(err).Fatal("sync service unexpectedly closed")
	}
}

func mustGetDB() *sql.DB {
	db, err := sql.Open("postgres", opts.Postgres)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create postgres connection")
	}
	db.SetMaxOpenConns(opts.PostgresMaxOpenConnections)
	db.SetMaxIdleConns(opts.PostgresMaxIdleConnections)

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

func mustGetConsumer(s storage.Storage) consumer.Consumer {
	fetcher, err := ariadne.New(context.Background(), opts.BlockchainNode, opts.BlockchainTimeout)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create blocks fetcher")
	}

	return blockchain.New(fetcher, s, opts.BlockchainRetryInterval, opts.BlockchainLastBlockRetryInterval)
}
