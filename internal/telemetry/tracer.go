package telemetry

import (
	"context"
	"net/url"

	"github.com/mdobak/go-xerrors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var ErrInitTrace = xerrors.Message("failed initializing tracer")

type (
	tracerConfig struct {
		endpoint string
		opts     []otlptracehttp.Option
	}
	TracerOption interface {
		apply(tracerConfig) tracerConfig
	}
	TracerOptionFunc func(tracerConfig) tracerConfig

	TracerProvider struct {
		*trace.TracerProvider
		endpoint string
	}
)

func (tof TracerOptionFunc) apply(tc tracerConfig) tracerConfig { return tof(tc) }

func WithEndpoint(endpoint string) TracerOption {
	return TracerOptionFunc(func(tc tracerConfig) tracerConfig {
		tc.endpoint = endpoint
		return tc
	})
}

func WithHeaders(headers map[string]string) TracerOption {
	return TracerOptionFunc(func(tc tracerConfig) tracerConfig {
		tc.opts = append(tc.opts, otlptracehttp.WithHeaders(headers))
		return tc
	})
}

func newTracerConfig(opts []TracerOption) (tc tracerConfig) {
	// https://github.com/open-telemetry/opentelemetry-go/blob/v1.31.0/exporters/otlp/otlptrace/otlptracehttp/internal/otlpconfig/options.go#L77
	tc.endpoint = "https://localhost:4318/v1/traces"
	for _, opt := range opts {
		tc = opt.apply(tc)
	}
	return
}

func NewTracerProvider(name string, opts ...TracerOption) (*TracerProvider, error) {
	config := newTracerConfig(opts)
	ctx := context.Background()
	traceOpts := append(config.opts, otlptracehttp.WithEndpointURL(config.endpoint))
	exp, err := otlptracehttp.New(ctx, traceOpts...)
	if err != nil {
		return nil, xerrors.WithWrapper(ErrInitTrace, err)
	}
	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(name)))
	if err != nil {
		return nil, xerrors.WithWrapper(ErrInitTrace, err)
	}
	provider := &TracerProvider{
		TracerProvider: trace.NewTracerProvider(trace.WithBatcher(exp), trace.WithResource(res)),
		endpoint:       config.endpoint,
	}

	// set as global to let NewSpan catch it
	otel.SetTracerProvider(provider)
	// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/31fe1e4559491449920a9eedd640aa7b6dce7086/instrumentation/net/http/httptrace/otelhttptrace/example/client/client.go#L43
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}))

	return provider, nil
}

func (tc *TracerProvider) Endpoint() string {
	endpointURL, _ := url.Parse(tc.endpoint)
	return endpointURL.Host
}
