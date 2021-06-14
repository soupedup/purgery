package middleware

import (
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/soupedup/purgery/internal/safe"

	"github.com/soupedup/purgery/internal/rest/internal/render"
)

// Auth implements a BasicAuth middleware.
func Auth(key string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _, ok := r.BasicAuth()

		if !ok || !safe.Compare(key, user) {
			render.Unauthorized(w)

			return
		}

		h.ServeHTTP(w, r)
	})
}

// Log wraps the Handler with logging.
func Log(logger *zap.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := fetchRecorder(w)
		defer rec.release()

		h.ServeHTTP(rec, r)

		logger.Info("processed.",
			zap.String("addr", r.RemoteAddr),
			zap.String("path", r.URL.Path),
			zap.Int("status", rec.status),
			zap.Duration("after", time.Since(rec.startedAt)),
		)
	})
}

type recorder struct {
	http.ResponseWriter
	startedAt time.Time
	status    int
}

func (rec *recorder) release() {
	rec.ResponseWriter = nil
	recorders.Put(rec)
}

func (rec *recorder) WriteHeader(status int) {
	rec.status = status
	rec.ResponseWriter.WriteHeader(status)
}

func fetchRecorder(w http.ResponseWriter) (rec *recorder) {
	rec = recorders.Get().(*recorder)
	rec.ResponseWriter = w
	rec.status = http.StatusOK
	rec.startedAt = time.Now()
	return
}

var recorders = sync.Pool{
	New: func() interface{} {
		return new(recorder)
	},
}
