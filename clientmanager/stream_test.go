package clientmanager_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/clientmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/stretchr/testify/assert"
)

func TestCallStream(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	t.Run("SSE streaming", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			flusher, ok := w.(http.Flusher)
			assert.True(t, ok)
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			for i := range 3 {
				_, _ = fmt.Fprintf(w, "data: chunk %d\n\n", i)
				flusher.Flush()
			}
		}))
		defer ts.Close()

		streamResp, err := clientmanager.CallStream(ctx, ts.URL)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, streamResp.StatusCode)
		assert.True(t, streamResp.IsSuccess())
		assert.NotEmpty(t, streamResp.Header.Get("Content-Type"))
		defer streamResp.Close()

		var chunks []string
		scanner := bufio.NewScanner(streamResp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if after, found := strings.CutPrefix(line, "data: "); found {
				chunks = append(chunks, after)
			}
		}
		assert.NoError(t, scanner.Err())
		assert.Equal(t, []string{"chunk 0", "chunk 1", "chunk 2"}, chunks)
	})

	t.Run("methods and options pass through", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if r.Header.Get("Authorization") != "Bearer test-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer ts.Close()

		streamResp, err := clientmanager.CallStream(ctx, ts.URL,
			clientmanager.WithMethod(http.MethodPost),
			clientmanager.WithAuth(clientmanager.AuthBearer("test-token")),
		)
		assert.NoError(t, err)
		defer streamResp.Close()
		assert.Equal(t, http.StatusNoContent, streamResp.StatusCode)
	})

	t.Run("error status code", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"server error"}`))
		}))
		defer ts.Close()

		streamResp, err := clientmanager.CallStream(ctx, ts.URL)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, streamResp.StatusCode)
		assert.False(t, streamResp.IsSuccess())
		defer streamResp.Close()

		body, _ := io.ReadAll(streamResp.Body)
		assert.Equal(t, `{"error":"server error"}`, string(body))
	})
}

func TestCallStreamCloseIdempotent(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data"))
	}))
	defer ts.Close()

	streamResp, err := clientmanager.CallStream(ctx, ts.URL)
	assert.NoError(t, err)

	err1 := streamResp.Close()
	err2 := streamResp.Close()
	// Both calls should succeed (no panic, no error from double-close)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
}

func TestCallStreamExecuteError(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	t.Run("invalid endpoint", func(t *testing.T) {
		res, err := clientmanager.CallStream(ctx, "://notfound")
		assert.Nil(t, res)
		assert.Error(t, err)
	})

	t.Run("unreachable host", func(t *testing.T) {
		res, err := clientmanager.CallStream(ctx, "http://127.0.0.1:1")
		assert.Nil(t, res)
		assert.Error(t, err)
	})
}

func TestCallBytes(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	t.Run("binary response", func(t *testing.T) {
		expected := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(expected)
		}))
		defer ts.Close()

		res, err := clientmanager.CallBytes(ctx, ts.URL)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.True(t, res.IsSuccess())
		assert.Equal(t, expected, res.Body)
		assert.Equal(t, expected, res.Raw)
	})

	t.Run("server error still returns raw bytes", func(t *testing.T) {
		expected := []byte(`{"error":"invalid"}`)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(expected)
		}))
		defer ts.Close()

		res, err := clientmanager.CallBytes(ctx, ts.URL)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		assert.False(t, res.IsSuccess())
		assert.Equal(t, expected, res.Body)
	})

	t.Run("server error with no body", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		res, err := clientmanager.CallBytes(ctx, ts.URL)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		assert.False(t, res.IsSuccess())
		assert.Empty(t, res.Body)
	})

	t.Run("empty response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer ts.Close()

		res, err := clientmanager.CallBytes(ctx, ts.URL)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, res.StatusCode)
		assert.Empty(t, res.Body)
	})

	t.Run("context timeout on read returns error", func(t *testing.T) {
		timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
		defer cancel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("a"))
			// The server has already sent headers + data; if the body reader
			// is slow or the context expires, io.ReadAll returns an error.
			// For a closed server connection we get a read error.
			<-r.Context().Done()
		}))
		defer ts.Close()

		_, err := clientmanager.CallBytes(timeoutCtx, ts.URL)
		assert.Error(t, err)
	})
}

func TestWithBodyReader(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	t.Run("custom io.Reader body", func(t *testing.T) {
		var receivedContentType string

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedContentType = r.Header.Get("Content-Type")
			body, _ := io.ReadAll(r.Body)
			assert.Equal(t, "custom body content", string(body))
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		res, err := clientmanager.Call[any](
			ctx, ts.URL,
			clientmanager.WithMethod(http.MethodPost),
			clientmanager.WithBodyReader(
				strings.NewReader("custom body content"),
				"text/plain",
			),
		)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.True(t, res.IsSuccess())
		assert.Equal(t, "text/plain", receivedContentType)
	})

	t.Run("precedence over WithRequestBody", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			// BodyReader should win, not the JSON-marshalled request body
			assert.Equal(t, "from-reader", string(body))
			assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		res, err := clientmanager.Call[any](
			ctx, ts.URL,
			clientmanager.WithMethod(http.MethodPost),
			clientmanager.WithRequestBody(map[string]string{"key": "value"}),
			clientmanager.WithBodyReader(
				strings.NewReader("from-reader"),
				"text/plain",
			),
		)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("nil reader falls through to JSON", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assert.Contains(t, string(body), `"hello"`)
			assert.Contains(t, r.Header.Get("Content-Type"), "application/json")
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		// A nil io.Reader should not match in the bodyReader case
		res, err := clientmanager.Call[any](
			ctx, ts.URL,
			clientmanager.WithMethod(http.MethodPost),
			clientmanager.WithBodyReader(nil, "text/plain"),
			clientmanager.WithRequestBody(map[string]string{"key": "hello"}),
		)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("nil reader with no body falls through", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		// A nil io.Reader with no other body options — should send no body
		res, err := clientmanager.Call[any](
			ctx, ts.URL,
			clientmanager.WithBodyReader(nil, ""),
		)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}

func TestClientManagerCallStream(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`hello from stream`))
	}))
	defer ts.Close()

	cm := clientmanager.New[string]()
	streamResp, err := cm.CallStream(ctx, ts.URL)
	assert.NoError(t, err)
	assert.True(t, streamResp.IsSuccess())
	defer streamResp.Close()

	body, _ := io.ReadAll(streamResp.Body)
	assert.Equal(t, "hello from stream", string(body))
}

func TestClientManagerCallBytes(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte{0x01, 0x02, 0x03})
	}))
	defer ts.Close()

	cm := clientmanager.New[any]()
	res, err := cm.CallBytes(ctx, ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, res.Body)
}

func TestCallStreamWithHost(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		assert.True(t, ok)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		for i := range 2 {
			_, _ = fmt.Fprintf(w, "event: test\ndata: item %d\n\n", i)
			flusher.Flush()
		}
	}))
	defer ts.Close()

	// Using WithHost with empty endpoint to test the host option
	streamResp, err := clientmanager.CallStream(ctx, "",
		clientmanager.WithHost(ts.URL),
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, streamResp.StatusCode)
	defer streamResp.Close()

	scanner := bufio.NewScanner(streamResp.Body)
	lines := 0
	for scanner.Scan() {
		lines++
	}
	assert.NoError(t, scanner.Err())
	assert.Greater(t, lines, 0) // should have received event lines
}

func TestCallStreamEmptyBody(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	streamResp, err := clientmanager.CallStream(ctx, ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, streamResp.StatusCode)
	defer streamResp.Close()

	body, _ := io.ReadAll(streamResp.Body)
	assert.Empty(t, body)
}

