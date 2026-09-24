package otel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

// Supported OTLP transport protocol identifiers for ExporterConfig.Protocol.
const (
	ProtocolGRPC         = "grpc"
	ProtocolHTTPProtobuf = "http/protobuf"
)

// ExporterConfig holds the configuration for the OpenTelemetry exporter
type ExporterConfig struct {
	Endpoint    string
	Insecure    bool
	Headers     map[string]string
	ServiceName string
	Environment string

	// Protocol selects the OTLP transport: ProtocolGRPC (default) or
	// ProtocolHTTPProtobuf. Leave empty to auto-detect from the
	// OTEL_EXPORTER_OTLP_TRACES_PROTOCOL / OTEL_EXPORTER_OTLP_PROTOCOL
	// environment variables, then from the Endpoint's URL scheme
	// ("https://" or "http://" implies ProtocolHTTPProtobuf), defaulting
	// to ProtocolGRPC for backward compatibility.
	Protocol string

	// CertificatePath, when set, is the filesystem path to a PEM-encoded CA
	// certificate used to verify the OTLP collector's TLS certificate.
	// Setting it implies a secure (non-insecure) connection for both
	// transport protocols.
	CertificatePath string

	// FromEnv, when true, ignores Endpoint, Insecure, Headers and
	// CertificatePath and lets the underlying otlptracegrpc/otlptracehttp
	// exporter configure itself entirely from OTEL_EXPORTER_OTLP_*
	// environment variables (including OTEL_EXPORTER_OTLP_CERTIFICATE).
	FromEnv bool
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

// resolveProtocol determines the OTLP transport protocol to use, in order of
// precedence: an explicitly configured value, then
// OTEL_EXPORTER_OTLP_TRACES_PROTOCOL, then OTEL_EXPORTER_OTLP_PROTOCOL, then
// the endpoint's URL scheme ("https://"/"http://" implies
// ProtocolHTTPProtobuf), defaulting to ProtocolGRPC.
func resolveProtocol(configured, endpoint string) string {
	if p := normalizeProtocol(configured); p != "" {
		return p
	}
	if p := normalizeProtocol(os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL")); p != "" {
		return p
	}
	if p := normalizeProtocol(os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")); p != "" {
		return p
	}
	if endpoint != "" {
		if u, err := url.Parse(endpoint); err == nil {
			switch u.Scheme {
			case "https", "http":
				return ProtocolHTTPProtobuf
			}
		}
	}
	return ProtocolGRPC
}

// normalizeProtocol maps a raw protocol string (from an option or an OTEL_*
// env var) to a supported Protocol constant, or "" if empty/unrecognized.
// "http" and "http/json" are treated as ProtocolHTTPProtobuf since this
// package only implements the protobuf transport.
func normalizeProtocol(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case ProtocolGRPC:
		return ProtocolGRPC
	case ProtocolHTTPProtobuf, "http", "http/json":
		return ProtocolHTTPProtobuf
	default:
		return ""
	}
}

// loadCATLSConfig reads a PEM-encoded CA certificate file and returns a
// *tls.Config that trusts only that CA, for use with
// otlptracehttp.WithTLSClientConfig.
func loadCATLSConfig(caCertPath string) (*tls.Config, error) {
	pemBytes, err := os.ReadFile(caCertPath) // #nosec G304 -- caCertPath is an operator-supplied config value, not untrusted input
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate file %q: %w", caCertPath, err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemBytes) {
		return nil, fmt.Errorf("failed to parse CA certificate file %q: no valid PEM certificates found", caCertPath)
	}

	return &tls.Config{RootCAs: pool}, nil
}

// newGRPCTraceExporter builds an OTLP gRPC span exporter from config.
func newGRPCTraceExporter(ctx context.Context, config *ExporterConfig) (sdktrace.SpanExporter, error) {
	if config.FromEnv {
		return otlptracegrpc.New(ctx)
	}

	exporterOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.Endpoint),
	}

	switch {
	case config.CertificatePath != "":
		creds, err := credentials.NewClientTLSFromFile(config.CertificatePath, "")
		if err != nil {
			return nil, fmt.Errorf("failed to load OTel CA certificate %q: %w", config.CertificatePath, err)
		}
		exporterOpts = append(exporterOpts, otlptracegrpc.WithTLSCredentials(creds))
	case config.Insecure:
		exporterOpts = append(exporterOpts, otlptracegrpc.WithInsecure())
	}

	if len(config.Headers) > 0 {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithHeaders(config.Headers))
	}

	return otlptracegrpc.New(ctx, exporterOpts...)
}

// newHTTPTraceExporter builds an OTLP HTTP/Protobuf span exporter from config.
func newHTTPTraceExporter(ctx context.Context, config *ExporterConfig) (sdktrace.SpanExporter, error) {
	if config.FromEnv {
		return otlptracehttp.New(ctx)
	}

	exporterOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(config.Endpoint),
	}

	switch {
	case config.CertificatePath != "":
		tlsConfig, err := loadCATLSConfig(config.CertificatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load OTel CA certificate %q: %w", config.CertificatePath, err)
		}
		exporterOpts = append(exporterOpts, otlptracehttp.WithTLSClientConfig(tlsConfig))
	case config.Insecure:
		exporterOpts = append(exporterOpts, otlptracehttp.WithInsecure())
	}

	if len(config.Headers) > 0 {
		exporterOpts = append(exporterOpts, otlptracehttp.WithHeaders(config.Headers))
	}

	return otlptracehttp.New(ctx, exporterOpts...)
}

// NewExporter creates a new OpenTelemetry exporter with the given configuration
func NewExporter(config *ExporterConfig) (*Exporter, error) {
	if config == nil {
		config = DefaultExporterConfig()
	}

	// Create resource attributes
	res, err := resource.New(
		context.Background(),
		resource.WithFromEnv(),
		resource.WithAttributes(
			attribute.String("service.name", config.ServiceName),
			attribute.String("deployment.environment", config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP trace exporter, selecting the transport protocol
	ctx := context.Background()
	protocol := resolveProtocol(config.Protocol, config.Endpoint)

	var traceExporter sdktrace.SpanExporter
	switch protocol {
	case ProtocolHTTPProtobuf:
		traceExporter, err = newHTTPTraceExporter(ctx, config)
	default:
		traceExporter, err = newGRPCTraceExporter(ctx, config)
	}
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

	// Make sure trace context can be injected into / extracted from carriers.
	EnsureW3CPropagator()

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
