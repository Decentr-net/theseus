// Package middleware ...
package middleware

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/Decentr-net/theseus/internal/middleware/memory"
)

// Storage ...
type Storage interface {
	Get(key string) []byte
	Set(key string, content []byte, duration time.Duration)
}

// Cached ...
func Cached(ttl time.Duration, handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	storage := memory.NewStorage()

	return func(w http.ResponseWriter, r *http.Request) {
		content := storage.Get(r.RequestURI)
		if content != nil {
			_, _ = w.Write(content)
		} else {
			c := httptest.NewRecorder()
			handler(c, r)

			for k, v := range c.Header() {
				w.Header()[k] = v
			}

			w.WriteHeader(c.Code)
			content := c.Body.Bytes()

			storage.Set(r.RequestURI, content, ttl)

			_, _ = w.Write(content)
		}
	}
}
