package logmanager

import (
	"errors"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"time"
)

func (txn *TxnRecord) writeLog() {
	statusCode := txn.attrs.Value().GetInt(internal.AttributeResponseCode)
	if txn.isHttp() {
		if internal.HasErrorInternalFromHttpStatusCode(statusCode) {
			txn.error = errors.New("internal server error")
		}
		if internal.HasErrorBusinessFromHttpStatusCode(statusCode) {
			txn.error = errors.New("business error")
			txn.hasBusinessError = true
		}
	}

	te := txn.logger.WithFields(txn.extractValues())
	if txn.hasError() {
		if txn.hasBusinessError {
			te.Warn(internal.ErrorToString(txn.error))
			return
		}

		te.Error(internal.ErrorToString(txn.error))
		return
	}

	te.Info()
}

func (txn *TxnRecord) reset() {
	txn.start = time.Time{}
	txn.stop = time.Time{}
	txn.duration = 0
	txn.error = nil
	txn.wroteHeader = false
	txn.attrs = nil
	txn.hasBusinessError = false
	txn.skipRequest = false
	txn.skipResponse = false
}

func (txn *TxnRecord) extractValues() map[string]interface{} {
	values := map[string]interface{}{
		txn.traceIDKey: txn.traceID,
		"name":         txn.name,
		"type":         txn.txnType,
		"start":        txn.start.Format(time.RFC3339),
		"latency":      txn.duration.Milliseconds(),
		"service":      txn.service,
	}

	if len(txn.tags) > 0 {
		values["tags"] = txn.tags
	}

	headers := internal.Header{
		Data:           internal.ToMapString(txn.attrs.Value().Get(internal.AttributeRequestHeaders)),
		AllowedHeaders: txn.exposeHeaders,
		IsDebugMode:    txn.debug || txn.exposeAllHeader,
	}.FilterHeaders()

	if len(headers) > 0 {
		values[internal.AttributeRequestHeaders] = headers
	}

	if txn.attrs.Value().IsNotEmpty(internal.AttributeRequestQueryParam) {
		values[internal.AttributeRequestQueryParam] = txn.attrs.Value().Get(internal.AttributeRequestQueryParam)
	}

	skipItems := []string{
		txn.traceIDKey,
		"name",
		"type",
		"start",
		"latency",
		"service",
		"tags",
		internal.AttributeRequestHeaders,
		internal.AttributeResponseHeaders,
		internal.AttributeRequestContentLength,
		internal.AttributeRequestQueryParam,
	}
	for k, v := range txn.attrs.Value().Values() {
		var found bool
		for _, skipItem := range skipItems {
			if k == skipItem {
				found = true
			}
		}
		if found {
			continue
		}
		if txn.skipRequest && k == internal.AttributeRequestBody {
			values[k] = "*"
			continue
		}
		if txn.skipResponse && k == internal.AttributeResponseBody {
			values[k] = "*"
			continue
		}
		values[k] = v
	}

	return values
}
