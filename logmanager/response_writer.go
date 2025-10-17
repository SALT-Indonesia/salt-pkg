package logmanager

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"net/http"
)

type replacementResponseWriter struct {
	thd      *TxnRecord
	original http.ResponseWriter
}

func (rw *replacementResponseWriter) Header() http.Header {
	return rw.original.Header()
}

func (rw *replacementResponseWriter) Write(b []byte) (n int, err error) {
	hdr := rw.original.Header()
	n, err = rw.original.Write(b)
	headersJustWritten(rw.thd, http.StatusOK, hdr)
	bodyJustWritten(rw.thd, b)
	return
}

func (rw *replacementResponseWriter) WriteHeader(code int) {
	hdr := rw.original.Header()
	rw.original.WriteHeader(code)
	headersJustWritten(rw.thd, code, hdr)
}

func headersJustWritten(tx *TxnRecord, code int, _ http.Header) {
	if nil == tx {
		return
	}
	if tx.wroteHeader {
		return
	}
	tx.wroteHeader = true

	// ResponseHeaderAttributes(tx.attrs, hdr)
	internal.ResponseCodeAttribute(tx.attrs, code)
}

func bodyJustWritten(tx *TxnRecord, body []byte) {
	if nil == tx {
		return
	}

	internal.ResponseBodyAttribute(tx.attrs, body)
}
