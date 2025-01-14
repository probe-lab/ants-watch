// telemetry code is taken from dennis-tra/nebula
// repurposed for ants

package metrics

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/ipfs/go-log/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

var logger = log.Logger("telemetry")

const (
	MeterName  = "github.com/probe-lab/ants-watch"
	TracerName = "github.com/probe-lab/ants-watch"
)

type Telemetry struct {
	Tracer                  trace.Tracer
	AntsCountGauge          metric.Int64Gauge
	TrackedRequestsCounter  metric.Int64Counter
	BulkInsertCounter       metric.Int64Counter
	BulkInsertSizeHist      metric.Int64Histogram
	BulkInsertLatencyMsHist metric.Int64Histogram
	CacheHitCounter         metric.Int64Counter
}

func NewTelemetry(tp trace.TracerProvider, mp metric.MeterProvider) (*Telemetry, error) {
	meter := mp.Meter(MeterName)

	antsCountGauge, err := meter.Int64Gauge("ants_count", metric.WithDescription("Number of running ants"))
	if err != nil {
		return nil, fmt.Errorf("ants_count gauge: %w", err)
	}

	trackedRequestsCounter, err := meter.Int64Counter("tracked_requests_count", metric.WithDescription("Number requests tracked"))
	if err != nil {
		return nil, fmt.Errorf("tracked_requests_count gauge: %w", err)
	}

	bulkInsertCounter, err := meter.Int64Counter("bulk_insert_count", metric.WithDescription("Number of bulk inserts"))
	if err != nil {
		return nil, fmt.Errorf("bulk_insert_count gauge: %w", err)
	}

	bulkInsertSizeHist, err := meter.Int64Histogram("bulk_insert_size", metric.WithDescription("Size of bulk inserts"), metric.WithExplicitBucketBoundaries(0, 10, 50, 100, 500, 1000))
	if err != nil {
		return nil, fmt.Errorf("bulk_insert_size histogram: %w", err)
	}

	bulkInsertLatencyMsHist, err := meter.Int64Histogram("bulk_insert_latency", metric.WithDescription("Latency of bulk inserts (ms)"), metric.WithUnit("ms"))
	if err != nil {
		return nil, fmt.Errorf("bulk_insert_latency histogram: %w", err)
	}

	cacheHitCounter, err := meter.Int64Counter("cache_hit_count", metric.WithDescription("Number of cache hits"))
	if err != nil {
		return nil, fmt.Errorf("cache_hit_counter gauge: %w", err)
	}

	return &Telemetry{
		Tracer:                  tp.Tracer(TracerName),
		TrackedRequestsCounter:  trackedRequestsCounter,
		BulkInsertCounter:       bulkInsertCounter,
		BulkInsertSizeHist:      bulkInsertSizeHist,
		BulkInsertLatencyMsHist: bulkInsertLatencyMsHist,
		CacheHitCounter:         cacheHitCounter,
		AntsCountGauge:          antsCountGauge,
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
	logger.Debugln("Starting telemetry endpoint", "addr", addr)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(rw http.ResponseWriter, req *http.Request) {
		logger.Debugln("Responding to health check")
		if HealthStatus.Load() {
			rw.WriteHeader(http.StatusOK)
		} else {
			rw.WriteHeader(http.StatusServiceUnavailable)
		}
	})

	logger.Info("Starting prometheus server", "addr", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Warnln("Error serving prometheus", "err", err)
	}
}
