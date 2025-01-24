package main

import (
	"net/http"
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
	DefaultConfigPath = "d2s.toml"
	Name              = "d2s"
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
		}))

	ctx.FatalIfErrorf(run(conf))
}

func run(c *config.Configuration) error {
	logger := c.NewLogger()
	logger.Debug().Object("config", c).Msg("dumping config")

	opts := []server.ServerOption{
		server.WithLogger(logger),
		server.WithHost(c.Host),
		server.WithPort(c.Port),
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
	srv.Handle("/lorem", lorem.Index, server.WithCache)
	srv.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		// w.Write([]byte("I'm about to panic!")) // this will send a response 200 as we write to resp
		panic("some unknown reason")
	})
	srv.HandleFunc("/error", app.Error)
	srv.HandleFunc("/wait", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("starting wait\n"))
		time.Sleep(10 * time.Second)
		w.Write([]byte("ending wait\n"))
	})

	return srv.Start()
}
