package logmanager

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"github.com/stretchr/testify/assert"
)

// mockFlusher is a mock ResponseWriter that implements http.Flusher
type mockFlusher struct {
	httptest.ResponseRecorder
	flushCalled bool
}

func (m *mockFlusher) Flush() {
	m.flushCalled = true
}

// mockHijacker is a mock ResponseWriter that implements http.Hijacker
type mockHijacker struct {
	httptest.ResponseRecorder
	hijackCalled bool
}

func (m *mockHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	m.hijackCalled = true
	return nil, nil, nil
}

// mockPusher is a mock ResponseWriter that implements http.Pusher
type mockPusher struct {
	httptest.ResponseRecorder
	pushCalled bool
	pushTarget string
	pushOpts   *http.PushOptions
}

func (m *mockPusher) Push(target string, opts *http.PushOptions) error {
	m.pushCalled = true
	m.pushTarget = target
	m.pushOpts = opts
	return nil
}

// mockFullWriter implements all optional interfaces
type mockFullWriter struct {
	httptest.ResponseRecorder
	flushCalled  bool
	hijackCalled bool
	pushCalled   bool
	pushTarget   string
}

func (m *mockFullWriter) Flush() {
	m.flushCalled = true
}

func (m *mockFullWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	m.hijackCalled = true
	return nil, nil, nil
}

func (m *mockFullWriter) Push(target string, opts *http.PushOptions) error {
	m.pushCalled = true
	m.pushTarget = target
	return nil
}

func TestReplacementResponseWriter_Flush(t *testing.T) {
	t.Run("supports Flush when underlying writer implements Flusher", func(t *testing.T) {
		mock := &mockFlusher{}
		txn := &TxnRecord{
			wroteHeader: false,
			attrs:       internal.NewAttributes(),
		}
		rw := &replacementResponseWriter{
			thd:      txn,
			original: mock,
		}

		// Type assertion should succeed
		flusher, ok := (http.ResponseWriter(rw)).(http.Flusher)
		assert.True(t, ok, "replacementResponseWriter should implement http.Flusher")
		assert.NotNil(t, flusher)

		// Call Flush
		flusher.Flush()

		// Verify it was called on underlying writer
		assert.True(t, mock.flushCalled, "Flush should be called on underlying writer")
		assert.True(t, txn.wroteHeader, "Headers should be marked as written after Flush")
	})

	t.Run("does not panic when underlying writer does not implement Flusher", func(t *testing.T) {
		mock := httptest.NewRecorder()
		txn := &TxnRecord{
			wroteHeader: false,
			attrs:       internal.NewAttributes(),
		}
		rw := &replacementResponseWriter{
			thd:      txn,
			original: mock,
		}

		// Should not panic
		assert.NotPanics(t, func() {
			rw.Flush()
		})
	})

	t.Run("does not write headers twice if already written", func(t *testing.T) {
		mock := &mockFlusher{}
		txn := &TxnRecord{
			wroteHeader: true,
			attrs:       internal.NewAttributes(),
		}
		rw := &replacementResponseWriter{
			thd:      txn,
			original: mock,
		}

		// Call Flush when headers already written
		rw.Flush()

		// wroteHeader should still be true (not reset)
		assert.True(t, txn.wroteHeader)
		assert.True(t, mock.flushCalled)
	})
}

func TestReplacementResponseWriter_Hijack(t *testing.T) {
	t.Run("supports Hijack when underlying writer implements Hijacker", func(t *testing.T) {
		mock := &mockHijacker{}
		txn := &TxnRecord{
			wroteHeader: false,
			attrs:       internal.NewAttributes(),
		}
		rw := &replacementResponseWriter{
			thd:      txn,
			original: mock,
		}

		// Type assertion should succeed
		hijacker, ok := (http.ResponseWriter(rw)).(http.Hijacker)
		assert.True(t, ok, "replacementResponseWriter should implement http.Hijacker")
		assert.NotNil(t, hijacker)

		// Call Hijack
		conn, bufrw, err := hijacker.Hijack()

		// Verify it was called on underlying writer
		assert.NoError(t, err)
		assert.Nil(t, conn)   // mock returns nil
		assert.Nil(t, bufrw)  // mock returns nil
		assert.True(t, mock.hijackCalled, "Hijack should be called on underlying writer")
		assert.True(t, txn.wroteHeader, "Headers should be marked as written after Hijack")
	})

	t.Run("returns error when underlying writer does not implement Hijacker", func(t *testing.T) {
		mock := httptest.NewRecorder()
		txn := &TxnRecord{
			wroteHeader: false,
			attrs:       internal.NewAttributes(),
		}
		rw := &replacementResponseWriter{
			thd:      txn,
			original: mock,
		}

		// Call Hijack
		conn, bufrw, err := rw.Hijack()

		// Should return error
		assert.Error(t, err)
		assert.Nil(t, conn)
		assert.Nil(t, bufrw)
		assert.Contains(t, err.Error(), "not supported")
	})
}

func TestReplacementResponseWriter_Push(t *testing.T) {
	t.Run("supports Push when underlying writer implements Pusher", func(t *testing.T) {
		mock := &mockPusher{}
		txn := &TxnRecord{
			wroteHeader: false,
			attrs:       internal.NewAttributes(),
		}
		rw := &replacementResponseWriter{
			thd:      txn,
			original: mock,
		}

		// Type assertion should succeed
		pusher, ok := (http.ResponseWriter(rw)).(http.Pusher)
		assert.True(t, ok, "replacementResponseWriter should implement http.Pusher")
		assert.NotNil(t, pusher)

		// Call Push
		opts := &http.PushOptions{Method: "GET"}
		err := pusher.Push("/resource.css", opts)

		// Verify it was called on underlying writer
		assert.NoError(t, err)
		assert.True(t, mock.pushCalled, "Push should be called on underlying writer")
		assert.Equal(t, "/resource.css", mock.pushTarget)
		assert.Equal(t, opts, mock.pushOpts)
	})

	t.Run("returns ErrNotSupported when underlying writer does not implement Pusher", func(t *testing.T) {
		mock := httptest.NewRecorder()
		txn := &TxnRecord{
			wroteHeader: false,
			attrs:       internal.NewAttributes(),
		}
		rw := &replacementResponseWriter{
			thd:      txn,
			original: mock,
		}

		// Call Push
		err := rw.Push("/resource.css", nil)

		// Should return ErrNotSupported
		assert.Equal(t, http.ErrNotSupported, err)
	})
}

func TestReplacementResponseWriter_InterfacePreservation(t *testing.T) {
	t.Run("preserves all optional interfaces when underlying writer supports them", func(t *testing.T) {
		mock := &mockFullWriter{}
		txn := &TxnRecord{
			wroteHeader: false,
			attrs:       internal.NewAttributes(),
		}
		rw := &replacementResponseWriter{
			thd:      txn,
			original: mock,
		}

		// Check all interface assertions
		_, isFlusher := (http.ResponseWriter(rw)).(http.Flusher)
		_, isHijacker := (http.ResponseWriter(rw)).(http.Hijacker)
		_, isPusher := (http.ResponseWriter(rw)).(http.Pusher)

		assert.True(t, isFlusher, "Should implement http.Flusher")
		assert.True(t, isHijacker, "Should implement http.Hijacker")
		assert.True(t, isPusher, "Should implement http.Pusher")
	})

	t.Run("basic ResponseWriter interface is always implemented", func(t *testing.T) {
		mock := httptest.NewRecorder()
		txn := &TxnRecord{
			wroteHeader: false,
			attrs:       internal.NewAttributes(),
		}
		rw := &replacementResponseWriter{
			thd:      txn,
			original: mock,
		}

		// Basic interface should always work
		var w http.ResponseWriter = rw
		assert.NotNil(t, w)

		// Basic methods should work
		header := w.Header()
		assert.NotNil(t, header)

		w.WriteHeader(http.StatusOK)
		assert.True(t, txn.wroteHeader)

		n, err := w.Write([]byte("test"))
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
	})
}

func TestReplacementResponseWriter_FlushBeforeWrite(t *testing.T) {
	t.Run("Flush before Write writes headers with StatusOK", func(t *testing.T) {
		mock := &mockFlusher{
			ResponseRecorder: *httptest.NewRecorder(),
		}
		txn := &TxnRecord{
			wroteHeader: false,
			attrs:       internal.NewAttributes(),
		}
		rw := &replacementResponseWriter{
			thd:      txn,
			original: mock,
		}

		// Flush before any Write
		rw.Flush()

		assert.True(t, txn.wroteHeader, "Headers should be marked as written")
		assert.True(t, mock.flushCalled, "Flush should be called")
	})
}

func TestReplacementResponseWriter_HijackWritesProperStatusCode(t *testing.T) {
	t.Run("Hijack writes headers with StatusSwitchingProtocols", func(t *testing.T) {
		mock := &mockHijacker{
			ResponseRecorder: *httptest.NewRecorder(),
		}
		txn := &TxnRecord{
			wroteHeader: false,
			attrs:       internal.NewAttributes(),
		}
		rw := &replacementResponseWriter{
			thd:      txn,
			original: mock,
		}

		// Hijack should mark headers as written
		rw.Hijack()

		assert.True(t, txn.wroteHeader, "Headers should be marked as written")
		assert.True(t, mock.hijackCalled, "Hijack should be called")
	})
}
