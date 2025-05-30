package main

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/alecthomas/kong"

	"github.com/platipy-io/d2s/app"
	"github.com/platipy-io/d2s/app/lorem"
	"github.com/platipy-io/d2s/config"
	"github.com/platipy-io/d2s/internal/telemetry"
	"github.com/platipy-io/d2s/server"
)

var (
	// DefaultConfigPath is the default location the application will use to
	// find the configuration.
	DefaultConfigPath     = "d2s.toml"
	EnvironmentConfigPath = "PLATIPY_CONFIG"
	Name                  = "d2s"
)

func main() {
	conf := &config.Configuration{}

	ctx := kong.Parse(conf,
		kong.Name(Name),
		kong.Description("Start the service."),
		kong.UsageOnError(),
		kong.Bind(conf),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.WithBeforeResolve(func() error {
			if err := conf.ParseFile(DefaultConfigPath); err != nil {
				return err
			}
			if path := os.Getenv(EnvironmentConfigPath); path != "" {
				if err := conf.ParseFile(path); err != nil {
					return err
				}
			}
			return nil
		}),
	)

	ctx.FatalIfErrorf(run(conf))
}

func run(c *config.Configuration) error {
	logger := c.NewLogger()
	logger.Debug().Object("config", c).Msg("dumping config")

	if err := c.InitCookie(); err != nil {
		return err
	}
	if err := c.InitOAuth(); err != nil {
		return err
	} else if c.IsBypassAuth() {
		logger.Warn().Msg("authentication bypass activated")
	}
	cache, err := server.MiddlewareCache()
	if err != nil {
		return err
	}
	db, err := c.NewClient()
	if err != nil {
		return err
	}

	opts := []server.ServerOption{
		server.WithLogger(logger),
		server.WithHost(c.Host),
		server.WithPort(c.Port),
		server.WithErrorHandler(app.ErrorHandler),
		server.WithNotFoundHandler(app.NotFoundHandler),
		server.WithDatabase(db),
	}
	if c.Tracer.Enabled {
		provider, err := telemetry.NewTracerProvider(
			"d2s",
			c.Tracer.Opts()...)
		if err != nil {
			return err
		}
		opts = append(opts, server.WithTracerProvider(provider))
	}

	srv, err := server.NewServer(opts...)
	if err != nil {
		logger.Fatal().Stack().Err(err).Msg("failed to instanciate server")
	}
	base := srv.With(server.MiddlewareUser(app.ErrorHandler))
	base.Get("/", app.Index)
	base.Post("/", app.IndexPost)
	base.HandleFunc("/lorem", lorem.Index, cache)
	base.HandleFunc("/alert", app.Alert)
	base.HandleFunc("/panic", func(_ *server.Context) error {
		// w.Write([]byte("I'm about to panic!")) // this will send a response 200 as we write to resp
		panic("some unknown reason")
	})
	if c.IsBypassAuth() {
		base.HandleFunc("/auth/login", app.LoginBypass)
	} else {
		base.HandleFunc("/auth/login", app.Login)
		base.HandleFunc("/auth/callback", app.Callback)
	}
	base.HandleFunc("/auth/logout", app.Logout)
	base.HandleFunc("/error", func(ctx *server.Context) error {
		app.ErrorHandler(ctx, errors.New("something bad happened"))
		return nil
	})
	base.HandleFunc("/wait", func(ctx *server.Context) error {
		ctx.ResponseWriter.Write([]byte("starting wait\n"))
		time.Sleep(10 * time.Second)
		ctx.ResponseWriter.Write([]byte("ending wait\n"))
		return nil
	})

	srv.HandleStd("/*", http.FileServer(http.Dir(c.Public)))
	return srv.Start()
}
