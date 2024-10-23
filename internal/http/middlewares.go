package http

import (
	"net/http"

	"github.com/platipy-io/d2s/internal/log"
	"github.com/platipy-io/d2s/internal/telemetry"

	"github.com/mdobak/go-xerrors"
)

func MiddlewareRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer xerrors.Recover(func(err error) {
			w.WriteHeader(http.StatusInternalServerError)
			log.Ctx(ctx).Error().Ctx(ctx).Stack().Err(err).
				Msg("recovering from panic!")
		})
		next.ServeHTTP(w, r)
	})
}

var MiddlewareLogger = log.Middleware

var MiddlewareOpenTelemetry = telemetry.MiddlewareTracing

var MiddlewareMetrics = telemetry.MiddlewareMetrics()
