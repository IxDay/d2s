package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	cache "github.com/IxDay/http-cache"
	"github.com/IxDay/http-cache/adapter/memory"
	"github.com/go-chi/chi/v5"
	"github.com/heptiolabs/healthcheck"
	"github.com/mdobak/go-xerrors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/platipy-io/d2s/app"
	"github.com/platipy-io/d2s/app/lorem"
	"github.com/platipy-io/d2s/internal/log"
	"github.com/platipy-io/d2s/internal/telemetry"
)

var timeout = 30 * time.Second
var ErrStarting = xerrors.Message("failed starting")
var ErrStopping = xerrors.Message("failed stopping")
var ErrCache = xerrors.Message("failed to initialize cache")

type Middleware = func(http.Handler) http.Handler

type serverConfig struct {
	host           string
	port           int
	logger         log.Logger
	tracerProvider *telemetry.TracerProvider
}

func (sc serverConfig) addr() string {
	return sc.host + ":" + strconv.Itoa(sc.port)
}

// ServerOption applies a configuration option value to a Server.
type ServerOption interface {
	apply(serverConfig) serverConfig
}

type ServerOptionFunc func(serverConfig) serverConfig

func (fn ServerOptionFunc) apply(c serverConfig) serverConfig {
	return fn(c)
}

func newServerConfig(opts []ServerOption) serverConfig {
	sc := serverConfig{port: 8080, logger: log.Nop()}
	for _, opt := range opts {
		sc = opt.apply(sc)
	}
	return sc
}

func WithHost(host string) ServerOption {
	return ServerOptionFunc(func(sc serverConfig) serverConfig {
		sc.host = host
		return sc
	})
}

func WithLogger(logger log.Logger) ServerOption {
	return ServerOptionFunc(func(sc serverConfig) serverConfig {
		sc.logger = logger
		return sc
	})
}

func WithPort(port int) ServerOption {
	return ServerOptionFunc(func(sc serverConfig) serverConfig {
		sc.port = port
		return sc
	})
}

func WithTracerProvider(provider *telemetry.TracerProvider) ServerOption {
	return ServerOptionFunc(func(sc serverConfig) serverConfig {
		sc.tracerProvider = provider
		return sc
	})
}

func ListenAndServe(opts ...ServerOption) error {
	router := chi.NewRouter()
	errChan := make(chan error)
	health := healthcheck.NewHandler()
	config := newServerConfig(opts)
	logger := config.logger
	server := http.Server{Addr: config.addr(), Handler: router}
	middlewares := []Middleware{MiddlewareMetrics, MiddlewareLogger(logger), MiddlewareRecover}

	if config.tracerProvider != nil {
		tracerMiddleware := MiddlewareOpenTelemetry("server",
			otelhttp.WithTracerProvider(config.tracerProvider))
		endpoint := config.tracerProvider.Endpoint()
		health.AddReadinessCheck("tracer", healthcheck.TCPDialCheck(endpoint, 5*time.Second))
		middlewares = append([]Middleware{tracerMiddleware}, middlewares...)
	}

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		logger.Info().Msg("received interrupt, closing server...")
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		errChan <- xerrors.New(server.Shutdown(ctx))
		cancel()
		close(errChan)
	}()
	memcached, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000000),
	)
	if err != nil {
		return xerrors.New(ErrCache, err)
	}

	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(10*time.Minute),
		cache.ClientWithRefreshKey("opn"),
		cache.ClientWithExpiresHeader(),
		cache.ClientWithVary("Hx-Request"),
	)
	if err != nil {
		return xerrors.New(ErrCache, err)
	}
	router.HandleFunc("/live", health.LiveEndpoint)
	router.HandleFunc("/ready", health.ReadyEndpoint)
	router.Handle("/metrics", promhttp.Handler())

	router.Route("/", func(r chi.Router) {
		r.Use(middlewares...)
		r.HandleFunc("/", app.Index)
		r.Handle("/lorem", cacheClient.Middleware(lorem.Index))
		r.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
			// w.Write([]byte("I'm about to panic!")) // this will send a response 200 as we write to resp
			panic("some unknown reason")
		})
		r.HandleFunc("/wait", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("starting wait\n"))
			time.Sleep(10 * time.Second)
			w.Write([]byte("ending wait\n"))
		})
	})

	logger.Info().Msg("starting server on: " + server.Addr)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return xerrors.New(ErrStarting, err)
	}
	if err := xerrors.WithWrapper(ErrStopping, <-errChan); err != nil {
		return err
	}
	logger.Info().Msg("server stopped properly")
	return nil
}
