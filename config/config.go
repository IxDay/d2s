package config

import (
	"errors"
	"io"
	"os"
	"time"

	"github.com/platipy-io/d2s/internal/log"
	"github.com/platipy-io/d2s/internal/telemetry"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type (
	// Configuration hold the current fields to tune the application
	Configuration struct {
		Dev  bool
		Host string
		Port int
		Logger
		Tracer
	}

	Logger struct {
		Level string
	}

	Tracer struct {
		Enabled  bool
		Endpoint string
		Headers  map[string]string
	}
)

func New(paths ...string) (config Configuration, err error) {
	for _, path := range paths {
		viper.SetConfigFile(path)
		if err = viper.MergeInConfig(); err != nil && !errors.Is(err, os.ErrNotExist) {
			return
		}
	}

	err = viper.Unmarshal(&config)
	return
}

func (t Tracer) Opts() (opts []telemetry.TracerOption) {
	if t.Endpoint != "" {
		opts = append(opts, telemetry.WithEndpoint(t.Endpoint))
	}
	if len(t.Headers) != 0 {
		opts = append(opts, telemetry.WithHeaders(t.Headers))
	}
	return opts
}

func (c Configuration) NewLogger(level zerolog.Level, override bool) zerolog.Logger {
	var output io.Writer = os.Stdout

	if c.Dev {
		output = zerolog.ConsoleWriter{Out: os.Stdout}
		zerolog.ErrorStackMarshaler = log.MarshalStackDev
		if !override {
			level = zerolog.TraceLevel
		}
	} else {
		zerolog.ErrorStackMarshaler = log.MarshalStack
	}
	zerolog.TimeFieldFormat = time.RFC3339Nano

	return zerolog.New(output).Level(level).Hook(log.TracingHook{}).
		With().Timestamp().Logger()
}
