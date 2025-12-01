package logmanager

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

// WithMaskConfigs configures the application with a list of MaskConfigs for masking sensitive data in logs.
// Deprecated: Use WithMaskingConfig with JSONPath support instead.
func WithMaskConfigs(maskConfigs MaskConfigs) Option {
	return func(app *Application) {
		if len(maskConfigs) > 0 {
			app.maskConfigs = maskConfigs
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
