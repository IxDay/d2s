package log

import (
	"context"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type (
	Level       = zerolog.Level
	Logger      = zerolog.Logger
	TracingHook struct{}
)

const (
	DebugLevel = zerolog.DebugLevel
	InfoLevel  = zerolog.InfoLevel
	WarnLevel  = zerolog.WarnLevel
	ErrorLevel = zerolog.ErrorLevel
	FatalLevel = zerolog.FatalLevel
	PanicLevel = zerolog.PanicLevel
	NoLevel    = zerolog.NoLevel
	Disabled   = zerolog.Disabled
	TraceLevel = zerolog.TraceLevel
)

func (h TracingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	span := trace.SpanContextFromContext(e.GetCtx())
	if !span.TraceID().IsValid() {
		return
	}
	e.Str("span_id", span.SpanID().String())
	e.Str("trace_id", span.TraceID().String())
}

func Ctx(ctx context.Context) *Logger { return zerolog.Ctx(ctx) }
func Nop() Logger                     { return zerolog.Nop() }

func FnWrapper(ctx context.Context, msg string) func() {
	logger := Ctx(ctx)
	logger.Debug().Ctx(ctx).Msg("entering " + msg)
	return func() {
		logger.Debug().Ctx(ctx).Msg("leaving " + msg)
	}
}
