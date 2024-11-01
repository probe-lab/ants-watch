// telemetry code is taken from dennis-tra/nebula
// repurposed for ants

package metrics

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmeter "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.19.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/atomic"
)

const (
	MeterName  = "github.com/probe-lab/ants-watch"
	TracerName = "github.com/probe-lab/ants-watch"
)

type Telemetry struct {
	Tracer                 trace.Tracer
	CacheQueriesCount      metric.Int64Counter
	InsertRequestHistogram metric.Int64Histogram
}

func NewTelemetry(tp trace.TracerProvider, mp metric.MeterProvider) (*Telemetry, error) {
	meter := mp.Meter(MeterName)

	cacheQueriesCount, err := meter.Int64Counter("cache_queries", metric.WithDescription("Number of queries to the LRU caches"))
	if err != nil {
		return nil, fmt.Errorf("cache_queries counter: %w", err)
	}

	insertRequestHistogram, err := meter.Int64Histogram("insert_request_timing", metric.WithDescription("Histogram of database query times for request insertions"), metric.WithUnit("milliseconds"))
	if err != nil {
		return nil, fmt.Errorf("cache_queries counter: %w", err)
	}

	return &Telemetry{
		Tracer:                 tp.Tracer(TracerName),
		CacheQueriesCount:      cacheQueriesCount,
		InsertRequestHistogram: insertRequestHistogram,
	}, nil
}

// HealthStatus is a global variable that indicates the health of the current
// process. This could either be "hey, I've started crawling and everything
// works fine" or "hey, I'm monitoring the network and all good".
var HealthStatus = atomic.NewBool(false)

// NewMeterProvider initializes a new opentelemetry meter provider that exports
// metrics using prometheus. To serve the prometheus endpoint call
// [ListenAndServe] down below.
func NewMeterProvider() (metric.MeterProvider, error) {
	exporter, err := prometheus.New(prometheus.WithNamespace("ants-watch"))
	if err != nil {
		return nil, fmt.Errorf("new prometheus exporter: %w", err)
	}

	return sdkmeter.NewMeterProvider(sdkmeter.WithReader(exporter)), nil
}

// NewTracerProvider initializes a new tracer provider to send traces to the
// given host and port combination. If any of the two values is the zero value
// this function will return a no-op tracer provider which effectively disables
// tracing.
func NewTracerProvider(ctx context.Context, host string, port int) (trace.TracerProvider, error) {
	if host == "" || port == 0 {
		return noop.NewTracerProvider(), nil
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%d", host, port)),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("new otel trace exporter: %w", err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("ants-watch"),
		)),
	), nil
}

// ListenAndServe starts an HTTP server and exposes prometheus and pprof
// metrics. It also exposes a health endpoint that can be probed with
// `ants health`.
func ListenAndServe(host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)
	log.WithField("addr", addr).Debugln("Starting telemetry endpoint")

	// profile 1% of contention events
	runtime.SetMutexProfileFraction(1)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(rw http.ResponseWriter, req *http.Request) {
		log.Debugln("Responding to health check")
		if HealthStatus.Load() {
			rw.WriteHeader(http.StatusOK)
		} else {
			rw.WriteHeader(http.StatusServiceUnavailable)
		}
	})

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.WithError(err).Warnln("Error serving prometheus")
	}
}
