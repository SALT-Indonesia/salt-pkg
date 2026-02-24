package otel

import (
	"context"
	"testing"
	"time"
)

// TestDefaultExporterConfig tests the default exporter configuration
func TestDefaultExporterConfig(t *testing.T) {
	config := DefaultExporterConfig()

	if config.Endpoint != "localhost:4317" {
		t.Errorf("expected endpoint 'localhost:4317', got '%s'", config.Endpoint)
	}

	if !config.Insecure {
		t.Error("expected insecure to be true")
	}

	if config.ServiceName != "logmanager-service" {
		t.Errorf("expected service name 'logmanager-service', got '%s'", config.ServiceName)
	}

	if config.Environment != "development" {
		t.Errorf("expected environment 'development', got '%s'", config.Environment)
	}
}

// TestNewNoopExporter tests creating a noop exporter
func TestNewNoopExporter(t *testing.T) {
	exporter := NewNoopExporter()

	if exporter == nil {
		t.Fatal("expected exporter to be non-nil")
	}

	if exporter.IsEnabled() {
		t.Error("expected noop exporter to be disabled")
	}

	tracer := exporter.Tracer()
	if tracer == nil {
		t.Fatal("expected tracer to be non-nil")
	}

	if tracer.enabled {
		t.Error("expected noop tracer to be disabled")
	}
}

// TestNoopTracer tests creating a noop tracer
func TestNoopTracer(t *testing.T) {
	tracer := NewNoopTracer()

	if tracer == nil {
		t.Fatal("expected tracer to be non-nil")
	}

	if tracer.enabled {
		t.Error("expected noop tracer to be disabled")
	}

	// Test starting a span with noop tracer
	ctx := context.Background()
	span, _ := tracer.Start(ctx, "test-span", nil, SpanKindInternal, time.Now())

	if span == nil {
		t.Fatal("expected span to be non-nil")
	}

	if !span.IsNil() {
		t.Error("expected noop span to be nil")
	}

	// Test that calling methods on noop span doesn't panic
	span.End()
	span.SetAttributes()
	span.SetError(nil)
	span.SetName("new-name")

	if span.TraceID() != "" {
		t.Error("expected noop span to have empty trace ID")
	}

	if span.SpanID() != "" {
		t.Error("expected noop span to have empty span ID")
	}
}

// TestSpanNilChecks tests that span methods handle nil receiver gracefully
func TestSpanNilChecks(t *testing.T) {
	var span *Span

	// All these should not panic
	span.End()
	span.SetAttributes()
	span.SetError(nil)
	span.SetName("test")

	if span.TraceID() != "" {
		t.Error("expected nil span to have empty trace ID")
	}

	if span.SpanID() != "" {
		t.Error("expected nil span to have empty span ID")
	}

	if !span.IsNil() {
		t.Error("expected nil span to be nil")
	}
}

// TestTracerID tests the TracerID field on span
func TestTracerID(t *testing.T) {
	tracer := NewNoopTracer()
	ctx := context.Background()
	span, _ := tracer.Start(ctx, "test", nil, SpanKindInternal, time.Now())

	customTraceID := "custom-trace-123"
	span.SetTracerID(customTraceID)

	if span.TracerID() != customTraceID {
		t.Errorf("expected tracer ID '%s', got '%s'", customTraceID, span.TracerID())
	}
}

// TestSpanKinds tests that all span kinds are properly converted
func TestSpanKinds(t *testing.T) {
	tracer := NewTracer("test", nil, false)
	ctx := context.Background()

	kinds := []struct {
		kind     SpanKind
		expected string
	}{
		{SpanKindInternal, "internal"},
		{SpanKindServer, "server"},
		{SpanKindClient, "client"},
		{SpanKindProducer, "producer"},
		{SpanKindConsumer, "consumer"},
	}

	for _, test := range kinds {
		t.Run(test.expected, func(t *testing.T) {
			// This test just verifies the kind doesn't cause panic
			// Actual kind conversion is handled by otel SDK
			span, _ := tracer.Start(ctx, "test", nil, test.kind, time.Now())
			if span != nil {
				span.End()
			}
		})
	}
}

// TestExporterShutdown tests graceful shutdown
func TestExporterShutdown(t *testing.T) {
	exporter := NewNoopExporter()

	ctx := context.Background()
	err := exporter.Shutdown(ctx)

	if err != nil {
		t.Errorf("expected no error on shutdown, got %v", err)
	}

	// Test that shutdown on nil exporter doesn't panic
	var nilExporter *Exporter
	err = nilExporter.Shutdown(ctx)

	if err != nil {
		t.Errorf("expected no error on nil exporter shutdown, got %v", err)
	}
}
