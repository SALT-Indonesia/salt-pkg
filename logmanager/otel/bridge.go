package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// SpanKind represents the type of span
type SpanKind int

const (
	SpanKindInternal SpanKind = iota
	SpanKindServer
	SpanKindClient
	SpanKindProducer
	SpanKindConsumer
)

// Span is a wrapper around otel.Span for simplified interaction
type Span struct {
	span     trace.Span
	context  context.Context
	tracerID string // For tracking custom trace ID
}

// End completes the span
func (s *Span) End() {
	if s == nil || s.span == nil {
		return
	}
	s.span.End()
}

// SetAttributes sets attributes on the span
func (s *Span) SetAttributes(attrs ...attribute.KeyValue) {
	if s == nil || s.span == nil {
		return
	}
	s.span.SetAttributes(attrs...)
}

// SetError records an error on the span
func (s *Span) SetError(err error) {
	if s == nil || s.span == nil || err == nil {
		return
	}
	s.span.SetStatus(codes.Error, err.Error())
	s.span.RecordError(err)
}

// SetName sets the span name
func (s *Span) SetName(name string) {
	if s == nil || s.span == nil {
		return
	}
	s.span.SetName(name)
}

// TraceID returns the trace ID as a string
func (s *Span) TraceID() string {
	if s == nil || s.span == nil {
		return ""
	}
	return s.span.SpanContext().TraceID().String()
}

// SpanID returns the span ID as a string
func (s *Span) SpanID() string {
	if s == nil || s.span == nil {
		return ""
	}
	return s.span.SpanContext().SpanID().String()
}

// Context returns the span's context
func (s *Span) Context() context.Context {
	if s != nil && s.context != nil {
		return s.context
	}
	return context.Background()
}

// IsNil returns true if the span is nil or disabled
func (s *Span) IsNil() bool {
	return s == nil || s.span == nil
}

// SetTracerID sets the custom trace ID for correlation
func (s *Span) SetTracerID(traceID string) {
	if s != nil {
		s.tracerID = traceID
	}
}

// TracerID returns the custom trace ID
func (s *Span) TracerID() string {
	if s != nil {
		return s.tracerID
	}
	return ""
}

// Tracer wraps otel.Tracer for creating spans
type Tracer struct {
	tracer  trace.Tracer
	service string
	enabled bool
}

// Start creates a new span with the given name and options
func (t *Tracer) Start(ctx context.Context, name string, parent *Span, kind SpanKind, startTime time.Time) (*Span, context.Context) {
	if !t.enabled {
		return &Span{}, ctx
	}

	// Convert span kind
	var spanKind trace.SpanKind
	switch kind {
	case SpanKindServer:
		spanKind = trace.SpanKindServer
	case SpanKindClient:
		spanKind = trace.SpanKindClient
	case SpanKindProducer:
		spanKind = trace.SpanKindProducer
	case SpanKindConsumer:
		spanKind = trace.SpanKindConsumer
	default:
		spanKind = trace.SpanKindInternal
	}

	// Build span options
	opts := []trace.SpanStartOption{
		trace.WithTimestamp(startTime),
		trace.WithSpanKind(spanKind),
	}

	// Use parent context if available
	var spanContext context.Context
	if parent != nil && parent.context != nil {
		spanContext = parent.context
	} else {
		spanContext = ctx
	}

	ctx, span := t.tracer.Start(spanContext, name, opts...)

	return &Span{
		span:    span,
		context: ctx,
	}, ctx
}

// NewTracer creates a new Tracer wrapper
func NewTracer(service string, tracer trace.Tracer, enabled bool) *Tracer {
	return &Tracer{
		tracer:  tracer,
		service: service,
		enabled: enabled,
	}
}

// NewNoopTracer creates a disabled tracer that does nothing
func NewNoopTracer() *Tracer {
	return &Tracer{
		tracer:  otel.Tracer("noop"),
		service: "",
		enabled: false,
	}
}
