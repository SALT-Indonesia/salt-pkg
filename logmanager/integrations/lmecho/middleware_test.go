package lmecho_test

import (
	"bytes"
	"context"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmecho"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestMiddleware_MultipartFormData(t *testing.T) {
	tests := []struct {
		name               string
		hasFile            bool
		formFields         map[string]string
		expectedFields     map[string]interface{}
		expectedFileField  string
		expectedFileName   string
		shouldLogRequest   bool
		description        string
	}{
		{
			name:    "multipart form data with file upload",
			hasFile: true,
			formFields: map[string]string{
				"title":       "Tech Conference 2025",
				"description": "Annual tech conference",
				"location":    "Jakarta",
			},
			expectedFields: map[string]interface{}{
				"title":       "Tech Conference 2025",
				"description": "Annual tech conference",
				"location":    "Jakarta",
			},
			expectedFileField: "poster",
			expectedFileName:  "test-poster.txt",
			shouldLogRequest:  true,
			description:       "Should log form fields and file metadata for multipart/form-data with files",
		},
		{
			name:    "multipart form data without file upload",
			hasFile: false,
			formFields: map[string]string{
				"title":       "Workshop 2025",
				"description": "Coding workshop",
				"location":    "Bandung",
			},
			expectedFields: map[string]interface{}{
				"title":       "Workshop 2025",
				"description": "Coding workshop",
				"location":    "Bandung",
			},
			shouldLogRequest: true,
			description:      "Should log form fields for multipart/form-data without files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := logmanager.NewTestableApplication()
			app.ResetLoggedEntries()

			e := echo.New()
			e.Use(lmecho.Middleware(app.Application))

			e.POST("/v1/event", func(c echo.Context) error {
				err := c.Request().ParseMultipartForm(10 << 20)
				assert.NoError(t, err, "Should parse multipart form without error")

				txn := logmanager.FromContext(c.Request().Context())
				assert.NotNil(t, txn, "Transaction should exist in context")

				txn.SetWebRequest(c.Request())

				return c.JSON(http.StatusOK, map[string]interface{}{
					"status":  201,
					"message": "event created successfully",
				})
			})

			body, contentType := createMultipartFormRequest(t, tt.formFields, tt.hasFile, tt.expectedFileField, tt.expectedFileName)

			req, err := http.NewRequest(http.MethodPost, "/v1/event", body)
			assert.NoError(t, err, "Should create request without error")
			req.Header.Set("Content-Type", contentType)

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code, "Should return 200 OK")

			assert.Equal(t, 1, app.CountLoggedEntries(), "Should have exactly one logged entry")

			assert.True(t, app.HasLoggedField("trace_id"), "Should log trace_id field")
			assert.True(t, app.HasLoggedField("method"), "Should log method field")
			assert.True(t, app.HasLoggedField("url"), "Should log url field")
			assert.True(t, app.HasLoggedField("status"), "Should log status field")

			if tt.shouldLogRequest {
				assert.True(t, app.HasLoggedField("request"), "Should log request field for multipart/form-data")

				requestData := app.GetLoggedField("request")
				assert.NotNil(t, requestData, "Request data should not be nil")

				requestMap, ok := requestData.(map[string]interface{})
				assert.True(t, ok, "Request data should be a map")

				for fieldName, expectedValue := range tt.expectedFields {
					actualValue, exists := requestMap[fieldName]
					assert.True(t, exists, "Request should contain field: %s", fieldName)
					assert.Equal(t, expectedValue, actualValue, "Field %s should have correct value", fieldName)
				}

				if tt.hasFile {
					filesData, exists := requestMap["_files"]
					assert.True(t, exists, "Request should contain _files field when file is uploaded")
					assert.NotNil(t, filesData, "Files data should not be nil")

					var filesArray []interface{}
					switch v := filesData.(type) {
					case []interface{}:
						filesArray = v
					case []map[string]interface{}:
						filesArray = make([]interface{}, len(v))
						for i, file := range v {
							filesArray[i] = file
						}
					default:
						t.Fatalf("Unexpected type for _files: %T", filesData)
					}

					assert.Greater(t, len(filesArray), 0, "Files array should not be empty")

					firstFile, ok := filesArray[0].(map[string]interface{})
					assert.True(t, ok, "First file should be a map")
					assert.Equal(t, tt.expectedFileField, firstFile["field"], "File field name should match")
					assert.Equal(t, tt.expectedFileName, firstFile["filename"], "File name should match")
					assert.NotNil(t, firstFile["size"], "File size should be logged")
					assert.NotNil(t, firstFile["header"], "File header should be logged")
				}
			}

			assert.Equal(t, "POST", app.GetLoggedField("method"), "Should log POST method")
			assert.Equal(t, "/v1/event", app.GetLoggedField("url"), "Should log correct URL")
			assert.Equal(t, 200, app.GetLoggedField("status"), "Should log 200 status")
		})
	}
}

func createMultipartFormRequest(t *testing.T, formFields map[string]string, hasFile bool, fileFieldName, fileName string) (io.Reader, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, val := range formFields {
		err := writer.WriteField(key, val)
		assert.NoError(t, err, "Should write form field without error")
	}

	if hasFile {
		fileContent := "Test file content for " + fileName
		part, err := writer.CreateFormFile(fileFieldName, fileName)
		assert.NoError(t, err, "Should create form file without error")

		_, err = io.Copy(part, strings.NewReader(fileContent))
		assert.NoError(t, err, "Should write file content without error")
	}

	err := writer.Close()
	assert.NoError(t, err, "Should close writer without error")

	return body, writer.FormDataContentType()
}
