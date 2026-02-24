package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/attribute"
)

// ExporterConfig holds the configuration for the OpenTelemetry exporter
type ExporterConfig struct {
	Endpoint    string
	Insecure    bool
	Headers     map[string]string
	ServiceName string
	Environment string
}

// DefaultExporterConfig returns the default exporter configuration
func DefaultExporterConfig() *ExporterConfig {
	return &ExporterConfig{
		Endpoint:    "localhost:4317",
		Insecure:    true,
		Headers:     make(map[string]string),
		ServiceName: "logmanager-service",
		Environment: "development",
	}
}

// Exporter wraps the OpenTelemetry tracer provider
type Exporter struct {
	tracerProvider *sdktrace.TracerProvider
	tracer         *Tracer
	config         *ExporterConfig
	enabled        bool
}

// Tracer returns the tracer instance
func (e *Exporter) Tracer() *Tracer {
	if e == nil {
		return NewNoopTracer()
	}
	return e.tracer
}

// Shutdown gracefully shuts down the exporter
func (e *Exporter) Shutdown(ctx context.Context) error {
	if e == nil || e.tracerProvider == nil {
		return nil
	}
	return e.tracerProvider.Shutdown(ctx)
}

// IsEnabled returns true if the exporter is enabled
func (e *Exporter) IsEnabled() bool {
	return e != nil && e.enabled
}

// NewExporter creates a new OpenTelemetry exporter with the given configuration
func NewExporter(config *ExporterConfig) (*Exporter, error) {
	if config == nil {
		config = DefaultExporterConfig()
	}

	// Create resource attributes
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", config.ServiceName),
			attribute.String("deployment.environment", config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP trace exporter
	ctx := context.Background()
	exporterOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.Endpoint),
		otlptracegrpc.WithInsecure(),
	}

	// Add headers if provided
	if len(config.Headers) > 0 {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithHeaders(config.Headers))
	}

	traceExporter, err := otlptracegrpc.New(ctx, exporterOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Create tracer provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tracerProvider)

	// Create our tracer wrapper
	tracer := NewTracer(
		config.ServiceName,
		tracerProvider.Tracer("logmanager"),
		true,
	)

	return &Exporter{
		tracerProvider: tracerProvider,
		tracer:         tracer,
		config:         config,
		enabled:        true,
	}, nil
}

// NewNoopExporter creates a disabled exporter that does nothing
func NewNoopExporter() *Exporter {
	return &Exporter{
		tracerProvider: nil,
		tracer:         NewNoopTracer(),
		config:         DefaultExporterConfig(),
		enabled:        false,
	}
}
