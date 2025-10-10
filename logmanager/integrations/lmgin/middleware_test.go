package lmgin_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgin"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware_appNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.Use(lmgin.Middleware(nil))
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	// Serve the test request
	r.ServeHTTP(w, req)

	// Check the response code
	assert.Equal(t, http.StatusOK, w.Code)
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
			gin.SetMode(gin.TestMode)
			r := gin.Default()

			// Reset logged entries before each test
			tt.app.ResetLoggedEntries()

			// Apply the middleware
			r.Use(middleware(tt.contexts), lmgin.Middleware(tt.app.Application))

			// Create a test route
			r.GET("/test", func(c *gin.Context) {
				value, exists := c.Get(string(tt.app.TraceIDContextKey()))

				// Assert in the context of the handler
				assert.True(t, exists)
				assert.NotEmpty(t, value)
				if !tt.randomTraceID {
					assert.Equal(t, tt.wantTraceID, value)
				}

				c.JSON(http.StatusOK, gin.H{
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

// TestMiddleware_TransactionInRequestContext tests that the transaction is accessible from c.Request.Context()
func TestMiddleware_TransactionInRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	app := logmanager.NewTestableApplication()
	app.ResetLoggedEntries()

	// Apply the middleware
	r.Use(lmgin.Middleware(app.Application))

	// Create a test route that simulates a service layer call
	r.GET("/test", func(c *gin.Context) {
		// This simulates calling a service/repository layer that only has access to context.Context
		ctx := c.Request.Context()

		// The transaction should be accessible from the request context
		tx := logmanager.FromContext(ctx)
		assert.NotNil(t, tx, "Transaction should be accessible from c.Request.Context()")

		// Also verify it's the same transaction stored in Gin context
		txFromGin, exists := c.Get(logmanager.TransactionContextKey.String())
		assert.True(t, exists, "Transaction should exist in Gin context")
		assert.Equal(t, txFromGin, tx, "Transaction from context.Context should match transaction from Gin context")

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Create a test HTTP request
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	w := httptest.NewRecorder()

	// Serve the test request
	r.ServeHTTP(w, req)

	// Check the response code
	assert.Equal(t, http.StatusOK, w.Code)
}

func middleware(contexts map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for k, v := range contexts {
			c.Set(k, v)
		}

		c.Next()
	}
}
