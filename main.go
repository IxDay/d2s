package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/platipy-io/d2s/config"
	"github.com/platipy-io/d2s/internal/http"
	"github.com/platipy-io/d2s/internal/log"
	"github.com/platipy-io/d2s/internal/telemetry"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	// DefaultConfigPath is the default location the application will use to
	// find the configuration.
	DefaultConfigPath = "d2s.toml"

	command = &cobra.Command{
		Use:   "d2s [flags]",
		Short: ".",
		RunE: func(cmd *cobra.Command, args []string) error {
			// https://github.com/spf13/cobra/issues/340
			cmd.SilenceUsage = true
			return run()
		},
	}
	configs      = []string{}
	logLevel     = config.LogLevel{Level: log.InfoLevel}
	logLevelFlag *pflag.Flag
)

func init() {
	flags := command.PersistentFlags()
	flags.String("host", "", "Host to listen to")
	flags.Int("port", 8080, "Port to listen to")
	flags.Bool("dev", false, "Activate dev mode")

	flags.StringArrayVar(&configs, "config", []string{DefaultConfigPath},
		"Path to a configuration file")

	logLevelFlag = flags.VarPF(&logLevel, "level", "",
		"Specify logger level; allowed: "+config.LogLevelsStr)

	_ = viper.BindPFlag("host", flags.Lookup("host"))
	_ = viper.BindPFlag("port", flags.Lookup("port"))
	_ = viper.BindPFlag("dev", flags.Lookup("dev"))
	_ = viper.BindPFlag("logger.level", logLevelFlag)
}

func main() {
	if err := command.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func run() error {
	conf, err := config.New(configs...)
	if err != nil {
		return err
	}
	logger := conf.NewLogger(logLevel.Level, logLevelFlag.Changed)
	logger.Debug().Object("config", conf).Msg("dumping config")

	opts := []http.ServerOption{http.WithLogger(logger), http.WithHost(conf.Host), http.WithPort(conf.Port)}
	if conf.Tracer.Enabled {
		provider, err := telemetry.NewTracerProvider("d2s", conf.Tracer.Opts()...)
		if err != nil {
			return err
		}
		opts = append(opts, http.WithTracerProvider(provider))
	}

	err = http.ListenAndServe(opts...)
	if errors.Is(err, http.ErrStopping) {
		logger.Error().Msg("failed to stop server")
	} else if errors.Is(err, http.ErrStarting) {
		logger.Fatal().Msg("failed to start server")
	}
	return err
}
