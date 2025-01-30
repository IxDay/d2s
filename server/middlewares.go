package server

import (
	"errors"
	"net/http"
	"time"

	cache "github.com/IxDay/http-cache"
	"github.com/IxDay/http-cache/adapter/memory"
	"github.com/platipy-io/d2s/internal/log"
	"github.com/platipy-io/d2s/internal/telemetry"

	"github.com/mdobak/go-xerrors"
)

func MiddlewareRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer xerrors.Recover(func(err error) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			log.Ctx(ctx).Error().Ctx(ctx).Stack().Err(err).
				Msg("recovering from panic!")
		})
		next.ServeHTTP(w, r)
	})
}

func MiddlewareUser(errHandler func(*Context, error)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := GetCookieUser(r)
			switch {
			case err == nil:
				r = SetUser(r, user)
			case errors.Is(err, http.ErrNoCookie):
				// pass without setting anything
			case errors.Is(err, ErrInvalidValue):
				// maybe 400
				errHandler(NewContext(w, r), err)
			default:
				ctx := r.Context()
				log.Ctx(ctx).Error().Ctx(ctx).Err(err).Msg("uncaught error")
			}
			next.ServeHTTP(w, r)
		})
	}
}

var ErrCache = xerrors.Message("failed to initialize cache")

func MiddlewareCache() (Middleware, error) {
	adapter, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000000),
	)
	if err != nil {
		return nil, xerrors.New(ErrCache, err)
	}

	client, err := cache.NewClient(
		cache.ClientWithAdapter(adapter),
		cache.ClientWithTTL(10*time.Minute),
		cache.ClientWithRefreshKey("opn"),
		cache.ClientWithExpiresHeader(),
		cache.ClientWithVary("Hx-Request"),
	)
	if err != nil {
		return nil, xerrors.New(ErrCache, err)
	}
	return client.Middleware, nil
}

var MiddlewareLogger = log.Middleware

var MiddlewareOpenTelemetry = telemetry.MiddlewareTracing

var MiddlewareMetrics = telemetry.MiddlewareMetrics()
