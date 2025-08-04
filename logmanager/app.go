package logmanager

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Application struct {
	name              string
	service           string
	debug             bool
	logger            *logrus.Logger
	logDir            string
	traceIDViaHeader  bool
	traceIDContextKey ContextKey
	traceIDHeaderKey  string
	maskConfigs       MaskConfigs
	maskingConfigs    []MaskingConfig
	tags              []string
	exposeHeaders     []string
	traceIDKey        string
}

// Service returns the service name used within the Application instance.
func (app *Application) Service() string {
	return app.service
}

// TraceIDKey returns the trace ID key used within the Application instance.
func (app *Application) TraceIDKey() string {
	return app.traceIDKey
}

// TraceIDContextKey returns the context key used to store the trace ID in the Application instance. If the receiver is nil, it returns an empty string.
func (app *Application) TraceIDContextKey() ContextKey {
	if nil == app {
		return ""
	}
	return app.traceIDContextKey
}

// TraceIDHeaderKey returns the header key used for trace ID identification within the Application instance.
// If the Application receiver is nil, it returns an empty string.
func (app *Application) TraceIDHeaderKey() string {
	if nil == app {
		return ""
	}
	return app.traceIDHeaderKey
}

// TraceIDViaHeader returns a boolean indicating if the trace ID should be extracted from the request header.
// If the Application receiver is nil, it returns false.
func (app *Application) TraceIDViaHeader() bool {
	if nil == app {
		return false
	}
	return app.traceIDViaHeader
}

// NewApplication creates a new Application instance with default settings, applying any provided configuration options.
// By default, the application's name is set to "default" and debugging mode is turned off.
// Options can be passed to customize the application, such as setting a custom name or enabling debugging.
// Once created, the application is assigned a logger that matches the specified debug level and log directory.
func NewApplication(opts ...Option) *Application {
	app := &Application{
		name:              "default",
		service:           "default",
		debug:             false, // Default Debug off
		traceIDContextKey: TraceIDContextKey,
		traceIDHeaderKey:  "X-Trace-Id",
		traceIDKey:        "trace_id",
		maskingConfigs: []MaskingConfig{
			{
				FieldPattern: "password",
				Type:         FullMask,
			},
		},
	}

	for _, opt := range opts {
		opt(app)
	}

	// Create a masker with a unified approach - convert all configs to a new format
	var allMaskingConfigs []internal.MaskingConfig

	// Convert legacy configs to a new format
	for _, legacyConfig := range app.maskConfigs.GetMaskConfigs() {
		allMaskingConfigs = append(allMaskingConfigs, internal.MaskingConfig{
			Field:     legacyConfig.Field, // Use Field for exact matching behavior
			Type:      legacyConfig.Type,
			ShowFirst: legacyConfig.ShowFirst,
			ShowLast:  legacyConfig.ShowLast,
		})
	}

	// Add new masking configs
	allMaskingConfigs = append(allMaskingConfigs, ConvertMaskingConfigs(app.maskingConfigs)...)

	masker := internal.NewJSONMasker(allMaskingConfigs)
	app.logger = newStandardLogger(app.debug, app.logDir, masker)

	return app
}

// StartHttp initializes a new HTTP transaction with the given trace ID and name.
// It returns a pointer to a Transaction object representing the HTTP transaction.
// If the Application receiver is nil, it returns a default Transaction with new attributes.
func (app *Application) StartHttp(traceID string, name string) *Transaction {
	if nil == app {
		return newEmptyTransaction()
	}

	return app.start(traceID, name, TxnTypeHttp)
}

// StartConsumer initializes a new consumer transaction with the specified trace ID.
// It returns a pointer to a Transaction object representing the consumer transaction.
// If the Application receiver is nil, it returns a default Transaction with new attributes.
// If the trace ID is empty, it generates a new UUID. Returns a pointer to a Transaction instance.
func (app *Application) StartConsumer(traceID string) *Transaction {
	if nil == app {
		return newEmptyTransaction()
	}

	return app.start(traceID, "consumer", TxnTypeConsumer)
}

// Start initializes a new transaction with the provided trace ID, name, and transaction type. Returns a pointer to Transaction.
func (app *Application) Start(traceID string, name string, transactionType TxnType) *Transaction {
	if nil == app {
		return newEmptyTransaction()
	}

	return app.start(traceID, name, transactionType)
}

func (app *Application) start(traceID string, name string, transactionType TxnType) *Transaction {
	if traceID == "" {
		traceID = uuid.NewString()
	}

	txn := &TxnRecord{
		name:          name,
		traceID:       traceID,
		txnType:       transactionType,
		start:         time.Now(),
		attrs:         internal.NewAttributes(),
		service:       app.name,
		logger:        app.logger,
		tags:          app.tags,
		exposeHeaders: app.exposeHeaders,
		debug:         app.debug,
		traceIDKey:    app.traceIDKey,
	}

	return &Transaction{
		TxnRecord:        txn,
		traceID:          traceID,
		txnRecords:       make(map[string]*TxnRecord),
		tags:             app.tags,
		traceIDKey:       app.traceIDKey,
		traceIDHeaderKey: app.traceIDHeaderKey,
	}
}
