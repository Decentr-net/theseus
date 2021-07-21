// Package health contains code for health checks.
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// nolint:gochecknoglobals
var (
	version = "dev"
	commit  = "undefined"
)

// GetVersion returns service's version and commit.
func GetVersion() string {
	return fmt.Sprintf("%s-%s", version, commit)
}

// VersionResponse ...
type VersionResponse struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// Pinger pings external service.
type Pinger interface {
	// Ping returns object with meta information and error
	Ping(ctx context.Context) (interface{}, error)
	// Name returns name of pinger
	Name() string
}

type subjectPinger struct {
	f func(ctx context.Context) error
	s string
}

// Ping ...
func (p subjectPinger) Ping(ctx context.Context) (interface{}, error) {
	return nil, p.f(ctx)
}

func (p subjectPinger) Name() string {
	return p.s
}

// SubjectPinger returns wrapper over Ping function which adds subject to error message.
// It is helpful for external Ping function, e.g. (sql.DB).Ping.
func SubjectPinger(s string, f func(ctx context.Context) error) Pinger {
	return subjectPinger{
		f: f,
		s: s,
	}
}

func Handler(timeout time.Duration, p ...Pinger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := context.WithTimeout(r.Context(), timeout) // nolint:govet
		gr, ctx := errgroup.WithContext(ctx)

		var mu sync.Mutex
		resp := struct {
			VersionResponse
			Meta   map[string]interface{} `json:"meta"`
			Errors map[string]error       `json:"errors"`
		}{
			VersionResponse: VersionResponse{Version: version, Commit: commit},
			Meta:            map[string]interface{}{},
			Errors:          map[string]error{},
		}

		for i := range p {
			v := p[i]
			gr.Go(func() error {
				m, err := v.Ping(ctx)
				if err != nil {
					logrus.WithError(err).Error("health check failed")
				}

				mu.Lock()
				resp.Meta[v.Name()] = m
				resp.Errors[v.Name()] = err
				mu.Unlock()

				return nil
			})
		}

		if err := gr.Wait(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		data, _ := json.Marshal(resp)
		w.Write(data)
	}
}
