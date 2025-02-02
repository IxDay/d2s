package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/heptiolabs/healthcheck"
	"github.com/mdobak/go-xerrors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/platipy-io/d2s/internal/log"
	"github.com/platipy-io/d2s/internal/telemetry"
)

var timeout = 30 * time.Second

type Middleware = func(http.Handler) http.Handler

type serverConfig struct {
	host            string
	port            int
	logger          log.Logger
	tracerProvider  *telemetry.TracerProvider
	errorHandler    func(*Context, error)
	notFoundHandler func(*Context)
}

func defaultErrorHandler(ctx *Context, err error) {
	ctx.Logger.Error().Ctx(ctx.Context()).Stack().Err(err).Msg("handling error")
	ctx.WriteHeader(http.StatusInternalServerError)
	ctx.ResponseWriter.Write([]byte(http.StatusText(http.StatusInternalServerError)))
}

func defaultNotFoundHandler(ctx *Context) {
	ctx.Logger.Warn().Ctx(ctx.Context()).Msg("page not found")
	ctx.WriteHeader(http.StatusNotFound)
	ctx.ResponseWriter.Write([]byte(http.StatusText(http.StatusNotFound)))
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

func WithNotFoundHandler(handler func(*Context)) ServerOption {
	return ServerOptionFunc(func(sc serverConfig) serverConfig {
		sc.notFoundHandler = handler
		return sc
	})
}

type (
	Server struct {
		server       *http.Server
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
	notFoundHandler := defaultNotFoundHandler
	errorHandler := defaultErrorHandler

	middlewares := []Middleware{
		MiddlewareMetrics, MiddlewareLogger(logger), MiddlewareRecover,
	}

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

	if config.notFoundHandler != nil {
		notFoundHandler = config.notFoundHandler
	}

	router.HandleFunc("/live", health.LiveEndpoint)
	router.HandleFunc("/ready", health.ReadyEndpoint)
	router.Handle("/metrics", promhttp.Handler())
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r)
		notFoundHandler(ctx)
		defer xerrors.Recover(func(err error) { errorHandler(ctx, err) })
	})

	return &Server{
		server:       &http.Server{Addr: config.addr(), Handler: router},
		router:       router.Route("/", func(r chi.Router) { r.Use(middlewares...) }),
		logger:       logger,
		errorHandler: errorHandler,
	}, nil
}

func (s *Server) stdHandler(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r)
		defer xerrors.Recover(func(err error) { s.errorHandler(ctx, err) })
		if err := handler.Handle(ctx); err != nil {
			s.errorHandler(ctx, err)
		}
	}
}

func (s *Server) Handle(pattern string, handler Handler, middlewares ...Middleware) {
	s.HandleStd(pattern, http.HandlerFunc(s.stdHandler(handler)), middlewares...)
}

func (s *Server) HandleFunc(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	s.Handle(pattern, handler, middlewares...)
}

func (s *Server) HandleStd(pattern string, handler http.Handler, middlewares ...Middleware) {
	if middlewares != nil {
		s.router.With(middlewares...).Handle(pattern, handler)
	} else {
		s.router.Handle(pattern, handler)
	}
}

func (s *Server) With(middlewares ...Middleware) *Server {
	return &Server{server: s.server, router: s.router.With(middlewares...),
		logger: s.logger, errorHandler: s.errorHandler}
}

func (s *Server) Get(pattern string, handler HandlerFunc) {
	s.router.Get(pattern, s.stdHandler(handler))
}

func (s *Server) Post(pattern string, handler HandlerFunc) {
	s.router.Post(pattern, s.stdHandler(handler))
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
