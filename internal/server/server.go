// Package server Theseus
//
// The Theseus is an off-chain service which provides access to community entities (posts, likes, follows)
//
//     Schemes: https
//     BasePath: /v1
//     Version: 1.2.1
//
//     Produces:
//     - application/json
//     Consumes:
//     - application/json
//
// swagger:meta
package server

import (
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"

	"github.com/Decentr-net/go-api"

	mm "github.com/Decentr-net/theseus/internal/middleware"
	"github.com/Decentr-net/theseus/internal/storage"
)

//go:generate swagger generate spec -t swagger -m -c . -o ../../static/swagger.json

const maxBodySize = 1024

type server struct {
	s storage.Storage
}

// SetupRouter setups handlers to chi router.
func SetupRouter(s storage.Storage, r chi.Router, timeout time.Duration) {
	r.Use(
		api.FileServerMiddleware("/docs", "static"),
		api.LoggerMiddleware,
		middleware.StripSlashes,
		cors.AllowAll().Handler,
		api.RequestIDMiddleware,
		api.RecovererMiddleware,
		api.TimeoutMiddleware(timeout),
		api.BodyLimiterMiddleware(maxBodySize),
	)

	srv := server{
		s: s,
	}

	r.Route("/v1", func(r chi.Router) {
		r.Get("/posts", srv.listPosts)
		r.Get("/posts/{owner}/{uuid}", srv.getPost)
		r.Get("/posts/{slug}", srv.getSharePostBySlug)
		r.Get("/profiles/stats", mm.Cached(10*time.Minute, srv.getDecentrStats))
		r.Get("/ddv/stats", mm.Cached(10*time.Minute, srv.getDDVStats))
		r.Get("/profiles/{address}/stats", srv.getProfileStats)
	})
}
