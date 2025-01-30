package telemetry

// Copied and adapted from https://github.com/766b/chi-prometheus
// Port https://github.com/zbindenren/negroni-prometheus for chi router
import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	bucketsLatency = []float64{300, 1200, 5000}
	bucketsSize    = []float64{200, 500, 900, 1500}
)

const (
	reqsName    = "http_requests_total"
	latencyName = "http_request_duration_milliseconds"
	sizeName    = "http_response_size_bytes"
)

func pattern(r *http.Request) (p string) {
	rctx := chi.RouteContext(r.Context())

	p = strings.Join(rctx.RoutePatterns, "")
	p = strings.Replace(p, "/*/", "/", -1)
	return
}

// Middleware returns a new prometheus Middleware handler that groups requests by the chi routing pattern.
// EX: /users/{firstName} instead of /users/bob
func MiddlewareMetrics() func(next http.Handler) http.Handler {
	reqs := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: reqsName,
			Help: "How many HTTP requests processed, partitioned by status code, method and HTTP path (with patterns).",
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(reqs)

	latency := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    latencyName,
		Help:    "How long it took to process the request, partitioned by status code, method and HTTP path (with patterns).",
		Buckets: bucketsLatency,
	},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(latency)

	size := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    sizeName,
			Help:    "A histogram of response sizes for requests.",
			Buckets: bucketsSize,
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(size)

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww, ok := w.(middleware.WrapResponseWriter)
			if !ok {
				ww = middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			}
			next.ServeHTTP(ww, r)

			labels := []string{r.Method, pattern(r)}
			switch status := ww.Status(); {
			case status < 200:
				labels = append(labels, "1XX")
			case status < 300:
				labels = append(labels, "2XX")
			case status < 400:
				labels = append(labels, "3XX")
			case status < 500:
				labels = append(labels, "4XX")
			default:
				labels = append(labels, "5XX")
			}

			reqs.WithLabelValues(labels...).Inc()
			latency.WithLabelValues(labels...).Observe(float64(time.Since(start).Nanoseconds()) / 1000000)
			size.WithLabelValues(labels...).Observe(float64(ww.BytesWritten()))
		}
		return http.HandlerFunc(fn)
	}
}

var MiddlewareTracing = otelhttp.NewMiddleware
