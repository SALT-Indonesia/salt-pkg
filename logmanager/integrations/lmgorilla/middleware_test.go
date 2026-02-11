package lmgorilla_test

import (
	"context"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgorilla"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware_appNil(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	lmgorilla.Middleware(nil).Middleware(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		app           *logmanager.TestableApplication
		contexts      map[logmanager.ContextKey]string
		headers       map[string]string
		randomTraceID bool
		wantTraceID   string
	}{
		{
			name: "it should be ok trace ID via header with request header",
			app: logmanager.NewTestableApplication(
				logmanager.WithTraceIDHeaderKey("X-Custom-ID"),
			),
			headers: map[string]string{
				"X-Custom-ID": "a",
			},
			wantTraceID: "a",
		},
		{
			name: "it should be ok trace ID via header with request header empty value",
			app: logmanager.NewTestableApplication(
				logmanager.WithTraceIDHeaderKey("X-Custom-ID"),
			),
			headers: map[string]string{
				"X-Custom-ID": "b",
			},
			wantTraceID: "b",
		},
		{
			name: "it should be ok trace ID via header without request header",
			app: logmanager.NewTestableApplication(
				logmanager.WithTraceIDHeaderKey("X-Custom-ID"),
			),
			randomTraceID: true,
		},
		{
			name: "it should be ok trace ID via context",
			app: logmanager.NewTestableApplication(
				logmanager.WithTraceIDContextKey("traceID"),
			),
			contexts: map[logmanager.ContextKey]string{
				"traceID": "c",
			},
			wantTraceID: "c",
		},
		{
			name: "it should be ok trace ID via context with empty value",
			app: logmanager.NewTestableApplication(
				logmanager.WithTraceIDContextKey("traceID"),
			),
			contexts: map[logmanager.ContextKey]string{
				"traceID": "",
			},
			randomTraceID: true,
		},
		{
			name: "it should be ok trace ID via context without context",
			app: logmanager.NewTestableApplication(
				logmanager.WithTraceIDContextKey("traceID"),
				logmanager.WithTags("test"),
			),
			randomTraceID: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := mux.NewRouter()

			// Reset logged entries before each test
			tt.app.ResetLoggedEntries()

			r.Use(middleware(tt.contexts), lmgorilla.Middleware(tt.app.Application))
			r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
				value, ok := r.Context().Value(tt.app.TraceIDContextKey()).(string)
				assert.True(t, ok)

				assert.NotEmpty(t, value)
				if !tt.randomTraceID {
					assert.Equal(t, tt.wantTraceID, value)
				}

				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"message":"ok"}`))
			}).Methods("GET")

			// Create a test HTTP request
			req, err := http.NewRequest(http.MethodGet, "/test", nil)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// Create a ResponseRecorder to record the response
			w := httptest.NewRecorder()

			// Serve the test request
			r.ServeHTTP(w, req)

			// Check the response code
			assert.Equal(t, http.StatusOK, w.Code)

			// Assert logged data keys and values
			assert.Equal(t, 1, tt.app.CountLoggedEntries(), "Should have exactly one logged entry")

			// Verify essential logged fields exist
			assert.True(t, tt.app.HasLoggedField("trace_id"), "Should log trace_id field")
			assert.True(t, tt.app.HasLoggedField("name"), "Should log name field")
			assert.True(t, tt.app.HasLoggedField("type"), "Should log type field")
			assert.True(t, tt.app.HasLoggedField("start"), "Should log start field")
			assert.True(t, tt.app.HasLoggedField("latency"), "Should log latency field")
			assert.True(t, tt.app.HasLoggedField("service"), "Should log service field")
			assert.True(t, tt.app.HasLoggedField("method"), "Should log method field")
			assert.True(t, tt.app.HasLoggedField("url"), "Should log url field")
			assert.True(t, tt.app.HasLoggedField("status"), "Should log status field")

			// Verify logged field values
			if !tt.randomTraceID {
				assert.Equal(t, tt.wantTraceID, tt.app.GetLoggedField("trace_id"), "Should log correct trace_id")
			} else {
				assert.NotEmpty(t, tt.app.GetLoggedField("trace_id"), "Should log non-empty trace_id")
			}

			assert.Equal(t, "GET /test", tt.app.GetLoggedField("name"), "Should log correct transaction name")
			assert.Equal(t, logmanager.TxnTypeHttp, tt.app.GetLoggedField("type"), "Should log HTTP transaction type")
			assert.Equal(t, "default", tt.app.GetLoggedField("service"), "Should log default service name")
			assert.Equal(t, "GET", tt.app.GetLoggedField("method"), "Should log HTTP method")
			assert.Equal(t, "/test", tt.app.GetLoggedField("url"), "Should log request URL")
			assert.Equal(t, 200, tt.app.GetLoggedField("status"), "Should log response status code")

			// Verify log level is Info for successful requests
			assert.Equal(t, logrus.InfoLevel, tt.app.GetLoggedLevel(), "Should log at Info level for successful requests")
			assert.Equal(t, "", tt.app.GetLoggedMessage(), "Should have empty message for Info level logs")
		})
	}
}

func middleware(contexts map[logmanager.ContextKey]string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for k, v := range contexts {
				req := r.Context()
				req = context.WithValue(req, k, v)
				r = r.WithContext(req)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func TestMiddleware_WithStreaming(t *testing.T) {
	app := logmanager.NewTestableApplication(
		logmanager.WithAppName("test-streaming"),
	)

	r := mux.NewRouter()
	r.Use(lmgorilla.Middleware(app.Application))

	// Handler that tests Flusher interface
	r.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		// Test that http.Flusher interface is available
		flusher, ok := w.(http.Flusher)
		assert.True(t, ok, "ResponseWriter should implement http.Flusher")

		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		// Send a few chunks
		for i := 0; i < 3; i++ {
			_, err := w.Write([]byte("data: test\n\n"))
			assert.NoError(t, err)
			flusher.Flush()
		}
	}).Methods(http.MethodGet)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	w := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "data: test")

	// Verify logging captured the response
	assert.Equal(t, 1, app.CountLoggedEntries(), "Should have logged one entry")
	assert.True(t, app.HasLoggedField("trace_id"), "Should log trace_id")
	assert.True(t, app.HasLoggedField("status"), "Should log status")
	assert.Equal(t, 200, app.GetLoggedField("status"), "Should log 200 status")
	assert.Equal(t, "GET /stream", app.GetLoggedField("name"), "Should log correct transaction name")
}

func TestMiddleware_WithHijacker(t *testing.T) {
	app := logmanager.NewTestableApplication(
		logmanager.WithAppName("test-hijacker"),
	)

	r := mux.NewRouter()
	r.Use(lmgorilla.Middleware(app.Application))

	// Handler that tests Hijacker interface availability
	r.HandleFunc("/hijack", func(w http.ResponseWriter, r *http.Request) {
		// Test that http.Hijacker interface type assertion works
		// Note: httptest.ResponseRecorder doesn't implement Hijacker,
		// so this will fail, but we're testing that the interface
		// is properly exposed (not that it succeeds)
		_, ok := w.(http.Hijacker)

		// In a real server, this would be true
		// In httptest, it's false, but the important part is
		// that the type assertion doesn't panic
		if ok {
			w.WriteHeader(http.StatusSwitchingProtocols)
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("hijacker not supported in test"))
		}
	}).Methods(http.MethodGet)

	req := httptest.NewRequest(http.MethodGet, "/hijack", nil)
	w := httptest.NewRecorder()

	// Should not panic
	assert.NotPanics(t, func() {
		r.ServeHTTP(w, req)
	})

	// Verify logging
	assert.Equal(t, 1, app.CountLoggedEntries(), "Should have logged one entry")
}

func TestMiddleware_WithPusher(t *testing.T) {
	app := logmanager.NewTestableApplication(
		logmanager.WithAppName("test-pusher"),
	)

	r := mux.NewRouter()
	r.Use(lmgorilla.Middleware(app.Application))

	// Handler that tests Pusher interface availability
	r.HandleFunc("/push", func(w http.ResponseWriter, r *http.Request) {
		// Test that http.Pusher interface type assertion works
		pusher, ok := w.(http.Pusher)

		if ok {
			// Try to push (will fail in test, but shouldn't panic)
			_ = pusher.Push("/resource.css", nil)
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("pusher not supported in test"))
		}
	}).Methods(http.MethodGet)

	req := httptest.NewRequest(http.MethodGet, "/push", nil)
	w := httptest.NewRecorder()

	// Should not panic
	assert.NotPanics(t, func() {
		r.ServeHTTP(w, req)
	})

	// Verify logging
	assert.Equal(t, 1, app.CountLoggedEntries(), "Should have logged one entry")
	assert.Equal(t, 200, app.GetLoggedField("status"), "Should log 200 status")
}
