package logmanager

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"github.com/sirupsen/logrus"
	"time"
)

type TxnRecord struct {
	name, traceID             string
	txnType                   TxnType
	start, stop               time.Time
	duration                  time.Duration
	error                     error
	wroteHeader               bool
	service                   string
	attrs                     *internal.Attributes
	logger                    *logrus.Logger
	hasBusinessError          bool
	tags                      []string
	debug                     bool
	exposeHeaders             []string
	traceIDKey                string
	skipRequest, skipResponse bool
	exposeAllHeader           bool
}

// Deprecated: MarkAsErrorBusiness is deprecated.
// Please use MarkAsBusinessError instead.
func (txn *TxnRecord) MarkAsErrorBusiness() {
	if nil == txn {
		return
	}
	txn.hasBusinessError = true
}

// MarkAsBusinessError marks the transaction as a business error.
// This method sets the transaction's state to a warning level while taking into account the current status code.
// It is typically used to differentiate between technical and business errors, where business errors are non-critical
// issues that do not prevent the completion of a process but might require attention and handling.
func (txn *TxnRecord) MarkAsBusinessError() {
	if nil == txn {
		return
	}
	txn.hasBusinessError = true
}

// SetBusinessError sets a business error on the transaction record.
// It marks the transaction with a business error state and records the provided error in the transaction.
func (txn *TxnRecord) SetBusinessError(err error) {
	if nil == txn {
		return
	}
	txn.MarkAsBusinessError()
	txn.error = err
}

// NoticeError sets the error for the transaction record and ends the transaction.
func (txn *TxnRecord) NoticeError(err error) {
	if nil == txn {
		return
	}

	txn.error = err
	txn.End()
}

// End marks the end time of the transaction, calculates its duration, and writes a log entry for the transaction record.
func (txn *TxnRecord) End() {
	if nil == txn || txn.start.IsZero() {
		return
	}

	txn.stop = time.Now()
	txn.duration = txn.stop.Sub(txn.start)

	txn.writeLog()
	txn.reset()
}

func (txn *TxnRecord) isHttp() bool {
	return txn.txnType == TxnTypeHttp
}

func (txn *TxnRecord) hasError() bool {
	return txn.error != nil || txn.hasBusinessError
}
