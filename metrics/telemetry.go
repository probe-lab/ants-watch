package metrics

import (
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	MeterName  = "github.com/probe-lab/ants-watch"
	TracerName = "github.com/probe-lab/ants-watch"
)

type Telemetry struct {
	Tracer                 trace.Tracer
	AntsCountGauge         metric.Int64Gauge
	TrackedRequestsCounter metric.Int64Counter
	DroppedRequestsCounter metric.Int64Counter
	ConnectCounter         metric.Int64Counter
	DisconnectCounter      metric.Int64Counter
}

// NewTelemetry builds the ants-watch instruments on the global OTel providers,
// which are configured by go-commons' tele.ServeMetrics / InitTraceProvider.
// Batch-insert metrics are emitted separately by db.BatchInserter.
func NewTelemetry() (*Telemetry, error) {
	meter := otel.GetMeterProvider().Meter(MeterName)

	antsCountGauge, err := meter.Int64Gauge("ants_count", metric.WithDescription("Number of running ants"))
	if err != nil {
		return nil, fmt.Errorf("ants_count gauge: %w", err)
	}

	trackedRequestsCounter, err := meter.Int64Counter("tracked_requests_count", metric.WithDescription("Number requests tracked"))
	if err != nil {
		return nil, fmt.Errorf("tracked_requests_count counter: %w", err)
	}

	droppedRequestsCounter, err := meter.Int64Counter("dropped_requests", metric.WithDescription("Number of requests dropped because of being duplicates"))
	if err != nil {
		return nil, fmt.Errorf("dropped_requests counter: %w", err)
	}

	connectCounter, err := meter.Int64Counter("connect", metric.WithDescription("Number of opened connections"))
	if err != nil {
		return nil, fmt.Errorf("connect counter: %w", err)
	}

	disconnectCounter, err := meter.Int64Counter("disconnect", metric.WithDescription("Number of closed connections"))
	if err != nil {
		return nil, fmt.Errorf("disconnect counter: %w", err)
	}

	return &Telemetry{
		Tracer:                 otel.GetTracerProvider().Tracer(TracerName),
		AntsCountGauge:         antsCountGauge,
		TrackedRequestsCounter: trackedRequestsCounter,
		DroppedRequestsCounter: droppedRequestsCounter,
		ConnectCounter:         connectCounter,
		DisconnectCounter:      disconnectCounter,
	}, nil
}
