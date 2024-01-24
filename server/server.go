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

	"github.com/platipy-io/d2s/internal/log"
	"github.com/platipy-io/d2s/internal/telemetry"
)

var timeout = 30 * time.Second
var ErrCache = xerrors.Message("failed to initialize cache")

type Middleware = func(http.Handler) http.Handler

type serverConfig struct {
	host           string
	port           int
	logger         log.Logger
	tracerProvider *telemetry.TracerProvider
	errorHandler   func(*Context, error)
}

func defaultErrorHandler(ctx *Context, err error) {
	ctx.Logger.Error().Ctx(ctx.Context()).Stack().Err(err).Msg("")
	ctx.WriteHeader(http.StatusInternalServerError)
	ctx.ResponseWriter.Write([]byte(http.StatusText(http.StatusInternalServerError)))
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

func WithErrorHandler(handler func(*Context, error)) ServerOption {
	return ServerOptionFunc(func(sc serverConfig) serverConfig {
		sc.errorHandler = handler
		return sc
	})
}

func newCache() (*cache.Client, error) {
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
	return client, nil
}

type (
	Server struct {
		server       *http.Server
		cache        *cache.Client
		router       chi.Router
		logger       log.Logger
		errorHandler func(*Context, error)
	}
)

func NewServer(opts ...ServerOption) (*Server, error) {
	health := healthcheck.NewHandler()
	config := newServerConfig(opts)
	router := chi.NewRouter()
	logger := config.logger
	errorHandler := defaultErrorHandler

	cache, err := newCache()
	if err != nil {
		return nil, err
	}

	middlewares := []Middleware{MiddlewareMetrics, MiddlewareLogger(logger), MiddlewareRecover}

	if config.tracerProvider != nil {
		tracerMiddleware := MiddlewareOpenTelemetry("server",
			otelhttp.WithTracerProvider(config.tracerProvider))
		endpoint := config.tracerProvider.Endpoint()
		health.AddReadinessCheck("tracer", healthcheck.TCPDialCheck(endpoint, 5*time.Second))
		middlewares = append([]Middleware{tracerMiddleware}, middlewares...)
	}

	if config.errorHandler != nil {
		errorHandler = config.errorHandler
	}

	router.HandleFunc("/live", health.LiveEndpoint)
	router.HandleFunc("/ready", health.ReadyEndpoint)
	router.Handle("/metrics", promhttp.Handler())

	return &Server{
		server:       &http.Server{Addr: config.addr(), Handler: router},
		router:       router.Route("/", func(r chi.Router) { r.Use(middlewares...) }),
		cache:        cache,
		logger:       logger,
		errorHandler: errorHandler,
	}, nil
}

func (s *Server) Handle(pattern string, handler Handler, opts ...HandlerOption) {
	config := newHandlerConfig(opts)
	var handlerStd http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{Logger: *log.Ctx(r.Context()), Request: r, ResponseWriter: w}
		if err := handler.Handle(ctx); err != nil {
			s.errorHandler(ctx, err)
		}
	})
	if config.cache {
		handlerStd = s.cache.Middleware(handlerStd)
	}

	s.router.Handle(pattern, handlerStd)
}

func (s *Server) HandleFunc(pattern string, handler HandlerFunc, opts ...HandlerOption) {
	s.Handle(pattern, handler, opts...)
}

func (s *Server) Start() error {
	errChan := make(chan error)

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		s.logger.Info().Msg("received interrupt, closing server...")
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		errChan <- xerrors.New(s.server.Shutdown(ctx))
		cancel()
		close(errChan)
	}()
	s.logger.Info().Msg("starting server on: " + s.server.Addr)
	err := s.server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		s.logger.Err(err).Msg("failed to start server")
		return err
	}
	if err := <-errChan; err != nil {
		s.logger.Err(err).Msg("failed to stop server")
		return err
	}
	s.logger.Info().Msg("server stopped properly")
	return nil
}

type handlerConfig struct {
	cache bool
}

// ServerOption applies a configuration option value to a Server.
type HandlerOption interface {
	apply(handlerConfig) handlerConfig
}

type HandlerOptionFunc func(handlerConfig) handlerConfig

func (fn HandlerOptionFunc) apply(c handlerConfig) handlerConfig {
	return fn(c)
}

func newHandlerConfig(opts []HandlerOption) handlerConfig {
	hc := handlerConfig{}
	for _, opt := range opts {
		hc = opt.apply(hc)
	}
	return hc
}

var WithCache = HandlerOptionFunc(func(hc handlerConfig) handlerConfig {
	hc.cache = true
	return hc
})
