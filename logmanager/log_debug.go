package logmanager

import (
	"context"
	"github.com/sirupsen/logrus"
)

// DebugWithContext logs a debug message with the trace ID from the context.
// It takes a context, a message string, and optional fields as parameters.
// If the context is nil, it logs the message without a trace ID.
// If the message is empty, it does nothing.
// Debug messages are only logged when debug mode is enabled in the application.
func DebugWithContext(ctx context.Context, msg string, fields ...map[string]string) {
	if msg == "" {
		return
	}

	var traceID string
	var logger *logrus.Logger
	var debugEnabled bool

	// Try to get the transaction from the context
	txn := FromContext(ctx)
	if txn != nil {
		// If we have a transaction, use its trace ID, logger, and debug setting
		traceID = txn.TraceID()
		if txn.TxnRecord != nil {
			if txn.TxnRecord.logger != nil {
				logger = txn.TxnRecord.logger
			}
			debugEnabled = txn.TxnRecord.debug
		}
	} else if ctx != nil {
		// If we don't have a transaction but have a context, try to get the trace ID directly
		if val, ok := ctx.Value(TraceIDContextKey.String()).(string); ok && val != "" {
			traceID = val
		}
		// Without a transaction, we can't determine debug mode, so we assume it's enabled
		// for backward compatibility
		debugEnabled = true
	} else {
		// No context provided, assume debug is enabled for backward compatibility
		debugEnabled = true
	}

	// Only proceed if debug mode is enabled
	if !debugEnabled {
		return
	}

	// If we don't have a logger from the transaction, create a new one
	if logger == nil {
		logger = logrus.New()
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	// Create the log entry with base fields
	logEntry := logger.WithField("type", "other")
	
	// Add trace ID if available
	if traceID != "" {
		logEntry = logEntry.WithField("trace_id", traceID)
	}

	// Add optional fields if provided
	if len(fields) > 0 && fields[0] != nil {
		for key, value := range fields[0] {
			logEntry = logEntry.WithField(key, value)
		}
	}

	// Log the debug message
	logEntry.Debug(msg)
}