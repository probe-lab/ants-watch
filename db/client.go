package db

import "context"

// RequestWriter buffers and persists DHT [Request]s. It is satisfied by
// go-commons' *db.BatchInserter[Request] (async batched inserts into
// ClickHouse) and by NoopWriter (used when no ClickHouse address is
// configured). Start must be called before Submit; Stop drains buffered rows.
type RequestWriter interface {
	Start(ctx context.Context)
	Submit(ctx context.Context, req Request) error
	Stop(ctx context.Context) error
}

// NoopWriter is a RequestWriter that discards every request. It is used when
// the queen runs without a configured ClickHouse backend.
type NoopWriter struct{}

var _ RequestWriter = (*NoopWriter)(nil)

func NewNoopWriter() *NoopWriter {
	return &NoopWriter{}
}

func (NoopWriter) Start(context.Context) {}

func (NoopWriter) Submit(context.Context, Request) error { return nil }

func (NoopWriter) Stop(context.Context) error { return nil }
