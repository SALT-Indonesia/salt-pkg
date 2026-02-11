package logmanager

import (
	"bufio"
	"errors"
	"net"
	"net/http"

	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
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

// Flush implements http.Flusher interface for streaming responses (SSE, chunked)
func (rw *replacementResponseWriter) Flush() {
	// Ensure headers are written before flushing
	if !rw.thd.wroteHeader {
		headersJustWritten(rw.thd, http.StatusOK, rw.original.Header())
	}

	// Delegate to underlying writer if it supports Flusher
	if flusher, ok := rw.original.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements http.Hijacker interface for WebSocket upgrades
func (rw *replacementResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	// Mark headers as written since hijacking bypasses normal flow
	if !rw.thd.wroteHeader {
		headersJustWritten(rw.thd, http.StatusSwitchingProtocols, rw.original.Header())
	}

	if hijacker, ok := rw.original.(http.Hijacker); ok {
		return hijacker.Hijack()
	}

	return nil, nil, errors.New("http.Hijacker interface not supported")
}

// Push implements http.Pusher interface for HTTP/2 server push
func (rw *replacementResponseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := rw.original.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
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
