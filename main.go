package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/pelletier/go-toml/v2"

	"github.com/platipy-io/d2s/app"
	"github.com/platipy-io/d2s/app/lorem"
	"github.com/platipy-io/d2s/config"
	"github.com/platipy-io/d2s/internal/telemetry"
	"github.com/platipy-io/d2s/server"
)

var (
	// DefaultConfigPath is the default location the application will use to
	// find the configuration.
	DefaultConfigPath = "d2s.toml"
	Name              = "d2s"
)

func exit(step string, err error) {
	fmt.Printf("failed to %s configuration file (%s): %s\n",
		step, DefaultConfigPath, err)
	os.Exit(1)
}

func main() {
	conf := &config.Configuration{}
	file, err := os.Open(DefaultConfigPath)
	if err == nil {
		if err := toml.NewDecoder(file).Decode(conf); err != nil {
			exit("unmarshal", err)
		}
	} else if !errors.Is(err, fs.ErrExist) {
		exit("read", err)
	}

	ctx := kong.Parse(conf,
		kong.Name(Name),
		kong.Description("Start the service."),
		kong.UsageOnError(),
		kong.Bind(conf),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}))

	ctx.FatalIfErrorf(run(conf))
}

func run(c *config.Configuration) error {
	if err := c.InitCookie(); err != nil {
		return err
	}
	if err := c.InitOAuth(); err != nil {
		return err
	}
	cache, err := server.MiddlewareCache()
	if err != nil {
		return err
	}
	logger := c.NewLogger()
	logger.Debug().Object("config", c).Msg("dumping config")

	opts := []server.ServerOption{
		server.WithLogger(logger),
		server.WithHost(c.Host),
		server.WithPort(c.Port),
		server.WithErrorHandler(app.ErrorHandler),
		server.WithNotFoundHandler(app.NotFoundHandler),
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
		logger.Fatal().Msg("failed to instanciate server")
	}
	srv.HandleFunc("/", app.Index)
	srv.HandleFunc("/lorem", lorem.Index, cache)
	srv.HandleFunc("/panic", func(_ *server.Context) error {
		// w.Write([]byte("I'm about to panic!")) // this will send a response 200 as we write to resp
		panic("some unknown reason")
	})
	if c.IsBypassAuth() {
		srv.HandleFunc("/auth/login", app.LoginBypass)
	} else {
		srv.HandleFunc("/auth/login", app.Login)
		srv.HandleFunc("/auth/callback", app.Callback)
	}
	srv.HandleFunc("/auth/logout", app.Logout)
	srv.HandleFunc("/error", func(ctx *server.Context) error {
		app.ErrorHandler(ctx, errors.New("something bad happened"))
		return nil
	})
	srv.HandleFunc("/wait", func(ctx *server.Context) error {
		ctx.ResponseWriter.Write([]byte("starting wait\n"))
		time.Sleep(10 * time.Second)
		ctx.ResponseWriter.Write([]byte("ending wait\n"))
		return nil
	})

	srv.HandleStd("/*", http.FileServer(http.Dir(c.Public)))
	return srv.Start()
}
