package logmanager

import (
	"context"
	"github.com/sirupsen/logrus"
)

// LogInfoWithContext logs an info message with the trace ID from the context.
// It takes a context, a message string, and optional fields as parameters.
// If the context is nil, it logs the message without a trace ID.
// If the message is empty, it does nothing.
func LogInfoWithContext(ctx context.Context, msg string, fields ...map[string]string) {
	if msg == "" {
		return
	}

	var traceID string
	var logger *logrus.Logger

	// Try to get the transaction from the context
	txn := FromContext(ctx)
	if txn != nil {
		// If we have a transaction, use its trace ID and logger
		traceID = txn.TraceID()
		if txn.TxnRecord != nil && txn.TxnRecord.logger != nil {
			logger = txn.TxnRecord.logger
		}
	} else if ctx != nil {
		// If we don't have a transaction but have a context, try to get the trace ID directly
		if val, ok := ctx.Value(TraceIDContextKey.String()).(string); ok && val != "" {
			traceID = val
		}
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

	// Log the info message
	logEntry.Info(msg)
}