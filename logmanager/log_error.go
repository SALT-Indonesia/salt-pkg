package logmanager

import (
	"context"
	"github.com/sirupsen/logrus"
)

// LogErrorWithContext logs an error with the trace ID from the context.
// It takes a context and an error as parameters.
// If the context is nil, it logs the error without a trace ID.
// If the error is nil, it does nothing.
func LogErrorWithContext(ctx context.Context, err error) {
	if err == nil {
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

	// Log the error with the trace ID if available
	if traceID != "" {
		logger.WithField("type", "other").WithField("trace_id", traceID).Error(err.Error())
	} else {
		logger.WithField("type", "other").Error(err.Error())
	}
}
