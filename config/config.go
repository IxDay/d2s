package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog"

	"github.com/platipy-io/d2s/internal/log"
	"github.com/platipy-io/d2s/internal/telemetry"
)

type (
	// Configuration hold the current fields to tune the application
	Configuration struct {
		Dev     Dev     `kong:"help='Activate dev mode',env='DEV'"`
		Configs Configs `kong:"help='Path to a configuration file (can be repeated)',name='config',sep='none',type='path'" toml:"-"`
		Host    string  `kong:"help='Host to listen to'"`
		Port    int     `kong:"help='Port to listen to',default='8080'"`
		Logger  `kong:"embed=''" toml:"logger"`
		Tracer  `kong:"-" toml:"tracer"`
	}

	Configs []string
	Dev     bool

	Logger struct {
		LogLevel Level `kong:"help='Set the logging level (debug|info|warn|error|fatal)',env='LOG_LEVEL',default='info'" toml:"level"`
	}

	Tracer struct {
		Enabled  bool
		Endpoint string
		Headers  map[string]string
	}
	Level struct {
		zerolog.Level
	}
)

func (l *Level) Decode(ctx *kong.DecodeContext) (err error) {
	if l.Level, err = zerolog.ParseLevel(ctx.Scan.Pop().String()); err != nil {
		return errors.New("invalid level")
	}
	return nil
}

func (c *Configs) BeforeResolve(ctx *kong.Context, trace *kong.Path, config *Configuration) error {
	configs := ctx.FlagValue(trace.Flag).(Configs)
	for _, file := range configs {
		if content, err := os.ReadFile(file); errors.Is(err, os.ErrNotExist) {
			fmt.Printf("config file: %s not found, skipping...\n", file)
		} else if err != nil {
			return err
		} else if err := toml.Unmarshal(content, config); err != nil {
			return err
		}
	}
	return nil
}

func (l *Level) BeforeResolve(ctx *kong.Context, trace *kong.Path, config *Configuration) error {
	if config.Dev {
		l.Level = zerolog.TraceLevel
	}
	return nil
}

func (d Dev) AfterApply(config *Configuration) error {
	if d {
		config.Logger.LogLevel.Level = zerolog.TraceLevel
	}
	return nil
}

func (l *Level) AfterApply(ctx *kong.Context, trace *kong.Path, config *Configuration) error {
	level := ctx.FlagValue(trace.Flag).(Level)
	l.Level = level.Level
	return nil
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

func (c Configuration) NewLogger() zerolog.Logger {
	var output io.Writer = os.Stdout
	var level zerolog.Level = c.Logger.LogLevel.Level

	if c.Dev {
		output = zerolog.ConsoleWriter{Out: os.Stdout}
		zerolog.ErrorStackMarshaler = log.MarshalStackDev
	} else {
		zerolog.ErrorStackMarshaler = log.MarshalStack
	}
	zerolog.TimeFieldFormat = time.RFC3339Nano

	return zerolog.New(output).Level(level).Hook(log.TracingHook{}).
		With().Timestamp().Logger()
}
