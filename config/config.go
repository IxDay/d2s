package config

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/mdobak/go-xerrors"
	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog"

	"github.com/platipy-io/d2s/data"
	"github.com/platipy-io/d2s/internal/github"
	"github.com/platipy-io/d2s/internal/log"
	"github.com/platipy-io/d2s/internal/telemetry"
	"github.com/platipy-io/d2s/server"
)

type (
	// Configuration hold the current fields to tune the application
	Configuration struct {
		Dev            Dev     `kong:"help='Activate dev mode',env='DEV'"`
		Configs        Configs `kong:"help='Path to a configuration file (can be repeated)',name='config',sep='none',type='path'" toml:"-"`
		Host           string  `kong:"help='Host to listen to'"`
		Port           int     `kong:"help='Port to listen to',default='8080'"`
		Public         string  `kong:"help='Path to public directory',default='./public'"`
		Authentication `kong:"-" toml:"authentication"`
		Logger         `kong:"embed=''" toml:"logger"`
		Tracer         `kong:"-" toml:"tracer"`
		Database       `kong:"-" toml:"database"`
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

	Database struct {
		Path string `toml:"path"`
	}

	Authentication struct {
		BypassToken  string `toml:"bypass-token"`
		Redirect     string
		ClientID     string `toml:"client-id"`
		ClientSecret string `toml:"client-secret"`
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
		if err := config.ParseFile(file); err != nil {
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

func (c Configuration) InitCookie() error {
	token, err := hex.DecodeString("13d6b4dff8f84a10851021ec8608f814570d562c92fe6b5ec4c9f595bcb3234b")
	if err != nil {
		return err
	}
	server.InitCookieStore(token)
	return nil
}

var ErrBypass = xerrors.Message("bypass can only be used with dev mode")

func (c Configuration) InitOAuth() error {
	if c.Authentication.BypassToken != "" {
		if c.Dev {
			return github.InitBypass(c.BypassToken)
		} else {
			return ErrBypass
		}
	}
	return github.InitOAuth(c.Redirect, c.ClientID, c.ClientSecret)
}

func (c *Configuration) ParseFile(path string) error {
	fmt.Println("reading", path)
	if content, err := os.ReadFile(path); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("config file: %s not found, skipping...\n", path)
	} else if err != nil {
		return xerrors.New("config file: ", path, "failed reading", err)
	} else if err := toml.Unmarshal(content, c); err != nil {
		return xerrors.New("config file: ", path, "failed parsing", err)
	}
	return nil
}

func (c Configuration) IsBypassAuth() bool {
	return bool(c.Dev) && c.Authentication.BypassToken != ""
}

func (d Database) NewClient() (*data.DB, error) {
	return data.NewDB(d.Path)
}
