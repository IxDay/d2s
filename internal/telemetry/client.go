package telemetry

import (
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const Timeout = 10 * time.Second
var HTTPClient = http.Client{
	Transport: otelhttp.NewTransport(http.DefaultTransport),
	Timeout: Timeout,
}
