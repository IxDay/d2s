package log

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

const (
	expireHeader = "Expires"
)

func handler(logger Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()
		ww, ok := w.(middleware.WrapResponseWriter)
		if !ok {
			ww = middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		}
		logger.Info().Ctx(ctx).
			Str("method", r.Method).
			Str("url", r.URL.Path).
			Str("user_agent", r.UserAgent()).
			Msg("starting request")

		if r.ContentLength != 0 {
			logger.Trace().Ctx(ctx).EmbedObject(Request(r)).Msg("dumping request")
		}
		next.ServeHTTP(ww, r.WithContext(logger.WithContext(ctx)))
		logger.Info().Ctx(ctx).
			Int("status", ww.Status()).
			Bool("cached", ww.Header().Get(expireHeader) != "").
			Int("size", ww.BytesWritten()).
			Dur("elapsed_ms", time.Since(start)).
			Msg("ending request")
	})
}

func Middleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return handler(logger, next)
	}
}
