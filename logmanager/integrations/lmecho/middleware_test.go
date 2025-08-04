package lmecho_test

import (
	"context"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmecho"
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware_appNil(t *testing.T) {
	e := echo.New()
	e.Use(lmecho.Middleware(nil))
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "ok",
		})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		app           *logmanager.TestableApplication
		contexts      map[string]string
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
			contexts: map[string]string{
				"traceID": "c",
			},
			wantTraceID: "c",
		},
		{
			name: "it should be ok trace ID via context with empty value",
			app: logmanager.NewTestableApplication(
				logmanager.WithTraceIDContextKey("traceID"),
			),
			contexts: map[string]string{
				"traceID": "",
			},
			randomTraceID: true,
		},
		{
			name: "it should be ok trace ID via context without context",
			app: logmanager.NewTestableApplication(
				logmanager.WithTraceIDContextKey("traceID"),
			),
			randomTraceID: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			// Reset logged entries before each test
			tt.app.ResetLoggedEntries()

			// Apply the middleware
			e.Use(middleware(tt.contexts), lmecho.Middleware(tt.app.Application))

			// Create a test route
			e.GET("/test", func(c echo.Context) error {
				value := c.Request().Context().Value(tt.app.TraceIDContextKey())

				// Assert in the context of the handler
				assert.NotNil(t, value)
				assert.NotEmpty(t, value)
				if !tt.randomTraceID {
					assert.Equal(t, tt.wantTraceID, value)
				}

				return c.JSON(http.StatusOK, map[string]string{
					"message": "ok",
				})
			})

			// Create a test HTTP request
			req, err := http.NewRequest(http.MethodGet, "/test", nil)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// Create a ResponseRecorder to record the response
			rec := httptest.NewRecorder()

			// Serve the test request
			e.ServeHTTP(rec, req)

			// Check the response code
			assert.Equal(t, http.StatusOK, rec.Code)

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

func middleware(contexts map[string]string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := req.Context()
			for k, v := range contexts {
				ctx = context.WithValue(ctx, logmanager.ContextKey(k), v)
			}
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}
