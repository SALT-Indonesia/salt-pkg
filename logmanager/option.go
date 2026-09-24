package logmanager

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/otel"
	"go.opentelemetry.io/otel/trace"
)

type Option func(*Application)

// WithDebug returns an Option that sets the debug mode to true for the Application.
func WithDebug() Option {
	return func(app *Application) {
		app.debug = true
	}
}

// WithAppName sets the application name for the Application.
func WithAppName(name string) Option {
	return func(app *Application) {
		if name != "" {
			app.name = name
		}
	}
}

// WithLogDir sets the log directory for the Application.
func WithLogDir(directory string) Option {
	return func(app *Application) {
		if directory != "" {
			app.logDir = directory
		}
	}
}

// WithTraceIDContextKey sets the key used to store the trace ID in the application's context.
func WithTraceIDContextKey(key ContextKey) Option {
	return func(app *Application) {
		if key != "" {
			app.traceIDContextKey = key
			app.traceIDViaHeader = false
		}
	}
}

// WithTraceIDHeaderKey sets the HTTP header key used for trace ID in the Application configuration.
func WithTraceIDHeaderKey(key string) Option {
	return func(app *Application) {
		if key != "" {
			app.traceIDHeaderKey = key
			app.traceIDViaHeader = true
		}
	}
}

// WithMaskingConfig configures the application with a list of MaskingConfig for masking sensitive data in logs using JSONPath.
func WithMaskingConfig(maskingConfigs []MaskingConfig) Option {
	return func(app *Application) {
		if len(maskingConfigs) > 0 {
			app.maskingConfigs = maskingConfigs
		}
	}
}

// WithTags assigns a list of tags to the Application instance through the provided Option functional argument.
func WithTags(tags ...string) Option {
	return func(app *Application) {
		app.tags = tags
	}
}

// WithSkipHeaders disables logging of request headers.
// This is useful in production environments where header logging generates excessive log volume.
func WithSkipHeaders() Option {
	return func(app *Application) {
		app.skipHeaders = true
	}
}

// WithExposeHeaders configures the Application to expose specified headers in the response or request by setting exposeHeaders.
func WithExposeHeaders(headers ...string) Option {
	return func(app *Application) {
		app.exposeHeaders = headers
	}
}

// WithTraceIDKey sets the trace ID key in the Application instance, if a non-empty key is provided.
func WithTraceIDKey(key string) Option {
	return func(app *Application) {
		if key != "" {
			app.traceIDKey = key
		}
	}
}

// WithService sets the service name for the Application.
func WithService(service string) Option {
	return func(app *Application) {
		if service != "" {
			app.service = service
		}
	}
}

// WithEnvironment sets the environment for the Application.
// When environment is set to "production", debug mode is automatically disabled
// unless explicitly set via WithDebug().
func WithEnvironment(environment string) Option {
	return func(app *Application) {
		if environment != "" {
			app.environment = environment
		}
	}
}

// WithSplitLevelOutput enables split-level log output routing following Twelve-Factor App principles.
// When enabled, log output is directed based on severity level:
//   - DEBUG, INFO, TRACE levels are written to os.Stdout
//   - WARN, ERROR, FATAL, PANIC levels are written to os.Stderr
//
// This is useful in containerized environments (Docker, Kubernetes) where log collectors
// treat stderr as an error state. Without this option, logrus defaults all output to stderr.
// This option is ignored when WithLogDir is set (file-based logging).
func WithSplitLevelOutput() Option {
	return func(app *Application) {
		app.splitLevelOutput = true
	}
}

// otelExporterConfig holds the configuration for OpenTelemetry exporter
type otelExporterConfig struct {
	endpoint        string
	insecure        bool
	headers         map[string]string
	serviceName     string
	protocol        string
	certificatePath string
	fromEnv         bool
}

// buildOTelConfig creates an otel.ExporterConfig from OTelExporterOption values
func buildOTelConfig(service, environment string, opts []OTelExporterOption) *otel.ExporterConfig {
	cfg := &otelExporterConfig{
		endpoint:    "localhost:4317",
		insecure:    true,
		headers:     make(map[string]string),
		serviceName: service,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &otel.ExporterConfig{
		Endpoint:        cfg.endpoint,
		Insecure:        cfg.insecure,
		Headers:         cfg.headers,
		ServiceName:     cfg.serviceName,
		Environment:     environment,
		Protocol:        cfg.protocol,
		CertificatePath: cfg.certificatePath,
		FromEnv:         cfg.fromEnv,
	}
}

// OTelExporterOption configures the OpenTelemetry exporter.
type OTelExporterOption func(*otelExporterConfig)

// WithOpenTelemetry enables OpenTelemetry trace export with custom configuration.
// When enabled, transactions will export spans to the configured OTLP endpoint.
// The exporter is initialized lazily on first transaction creation.
//
// Example usage:
//   app := logmanager.NewApplication(
//       logmanager.WithService("my-service"),
//       logmanager.WithOpenTelemetry(
//           logmanager.WithOTelEndpoint("localhost:4317"),
//           logmanager.WithOTelInsecure(),
//       ),
//   )
func WithOpenTelemetry(opts ...OTelExporterOption) Option {
	return func(app *Application) {
		app.otelEnabled = true
		app.otelExporterOptions = opts
	}
}

// WithOTelEndpoint sets the OTLP endpoint (default: localhost:4317).
func WithOTelEndpoint(endpoint string) OTelExporterOption {
	return func(cfg *otelExporterConfig) {
		if endpoint != "" {
			cfg.endpoint = endpoint
		}
	}
}

// WithOTelHeaders sets headers for OTLP connection (e.g., authentication).
func WithOTelHeaders(headers map[string]string) OTelExporterOption {
	return func(cfg *otelExporterConfig) {
		if headers != nil {
			cfg.headers = headers
		}
	}
}

// WithOTelServiceName sets the service name for OTel resource attributes.
// If not set, uses the service name from WithService().
func WithOTelServiceName(name string) OTelExporterOption {
	return func(cfg *otelExporterConfig) {
		if name != "" {
			cfg.serviceName = name
		}
	}
}

// WithOTelInsecure enables insecure connection (no TLS) for OTLP.
func WithOTelInsecure() OTelExporterOption {
	return func(cfg *otelExporterConfig) {
		cfg.insecure = true
	}
}

// WithOTelProtocol selects the OTLP transport protocol: "grpc" (default) or
// "http/protobuf". If not set, the protocol is auto-detected from the
// OTEL_EXPORTER_OTLP_PROTOCOL environment variable, then from the endpoint's
// URL scheme, defaulting to "grpc".
func WithOTelProtocol(protocol string) OTelExporterOption {
	return func(cfg *otelExporterConfig) {
		cfg.protocol = protocol
	}
}

// WithOTelCertificate sets the path to a PEM-encoded CA certificate used to
// verify the OTLP collector's TLS certificate. Setting this implies a secure
// connection, overriding WithOTelInsecure.
func WithOTelCertificate(caCertPath string) OTelExporterOption {
	return func(cfg *otelExporterConfig) {
		if caCertPath != "" {
			cfg.certificatePath = caCertPath
			cfg.insecure = false
		}
	}
}

// WithOTelFromEnv configures the OTel exporter purely from standard
// OTEL_EXPORTER_OTLP_* environment variables (OTEL_EXPORTER_OTLP_ENDPOINT,
// OTEL_EXPORTER_OTLP_PROTOCOL, OTEL_EXPORTER_OTLP_INSECURE,
// OTEL_EXPORTER_OTLP_HEADERS, OTEL_EXPORTER_OTLP_CERTIFICATE), ignoring any
// other WithOTel* options in the same WithOpenTelemetry call.
func WithOTelFromEnv() OTelExporterOption {
	return func(cfg *otelExporterConfig) {
		cfg.fromEnv = true
	}
}

// WithTracerProvider makes the Application reuse an existing OpenTelemetry
// TracerProvider instead of building its own exporter from the WithOTel*
// options.
//
// This is the recommended option when the host application already configures
// OpenTelemetry (its own provider, sampler, resource and exporters): logmanager
// spans are then emitted through that same provider, which is what allows a
// logmanager transaction to become part of the application's distributed trace
// rather than a separate one.
//
// Example:
//
//	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
//	app := logmanager.NewApplication(
//	    logmanager.WithService("my-service"),
//	    logmanager.WithTracerProvider(tp),
//	)
func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(app *Application) {
		if tp != nil {
			app.tracerProvider = tp
		}
	}
}
