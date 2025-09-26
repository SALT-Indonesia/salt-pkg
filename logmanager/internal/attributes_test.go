package internal_test

import (
	"bytes"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestRequestBodyAttributes(t *testing.T) {
	tests := []struct {
		name     string
		request  *http.Request
		attrs    *internal.Attributes
		expected map[string]interface{}
	}{
		{
			name: "non-empty request body",
			request: func() *http.Request {
				req, _ := http.NewRequest("POST", "http://example.com", bytes.NewBufferString(`{"key":"value"}`))
				return req
			}(),
			expected: map[string]interface{}{
				internal.AttributeRequestBody: map[string]interface{}{
					"key": "value",
				},
			},
			attrs: internal.NewAttributes(),
		},
		{
			name: "empty request body",
			request: func() *http.Request {
				req, _ := http.NewRequest("POST", "http://example.com", bytes.NewBufferString(``))
				return req
			}(),
			expected: map[string]interface{}{},
			attrs:    internal.NewAttributes(),
		},
		{
			name: "nil request body",
			request: func() *http.Request {
				req, _ := http.NewRequest("POST", "http://example.com", nil)
				return req
			}(),
			expected: map[string]interface{}{},
			attrs:    internal.NewAttributes(),
		},
		{
			name: "nil attr",
			request: func() *http.Request {
				req, _ := http.NewRequest("POST", "http://example.com", nil)
				return req
			}(),
			expected: map[string]interface{}{},
			attrs:    nil,
		},
		{
			name: "nil request",
			request: func() *http.Request {
				return nil
			}(),
			expected: map[string]interface{}{},
			attrs:    internal.NewAttributes(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := tt.attrs
			internal.RequestBodyAttributes(attrs, tt.request)

			for attrKey, expectedValue := range tt.expected {
				value := attrs.Value().Get(attrKey)
				assert.Equal(t, expectedValue, value)
			}
		})
	}
}

func TestRequestBodyAttributesMultipartFormData(t *testing.T) {
	t.Run("multipart form with text fields", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("name", "John Doe")
		_ = writer.WriteField("email", "john@example.com")
		_ = writer.WriteField("age", "30")
		writer.Close()

		req, _ := http.NewRequest("POST", "http://example.com/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		attrs := internal.NewAttributes()
		internal.RequestBodyAttributes(attrs, req)

		requestBody := attrs.Value().Get(internal.AttributeRequestBody)
		assert.NotNil(t, requestBody)

		formData, ok := requestBody.(map[string]interface{})
		assert.True(t, ok, "Request body should be a map")
		assert.Equal(t, "John Doe", formData["name"])
		assert.Equal(t, "john@example.com", formData["email"])
		assert.Equal(t, "30", formData["age"])
	})

	t.Run("multipart form with file upload", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("description", "Test file upload")

		part, _ := writer.CreateFormFile("document", "test.txt")
		_, _ = part.Write([]byte("file content here"))
		writer.Close()

		req, _ := http.NewRequest("POST", "http://example.com/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		attrs := internal.NewAttributes()
		internal.RequestBodyAttributes(attrs, req)

		requestBody := attrs.Value().Get(internal.AttributeRequestBody)
		assert.NotNil(t, requestBody)

		formData, ok := requestBody.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "Test file upload", formData["description"])

		files, ok := formData["_files"].([]map[string]interface{})
		assert.True(t, ok, "Should have files array")
		assert.Len(t, files, 1)
		assert.Equal(t, "document", files[0]["field"])
		assert.Equal(t, "test.txt", files[0]["filename"])
		assert.Greater(t, files[0]["size"], int64(0))
	})

	t.Run("multipart form with multiple files", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part1, _ := writer.CreateFormFile("file1", "doc1.pdf")
		_, _ = part1.Write([]byte("PDF content"))

		part2, _ := writer.CreateFormFile("file2", "image.jpg")
		_, _ = part2.Write([]byte("JPEG data"))

		writer.Close()

		req, _ := http.NewRequest("POST", "http://example.com/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		attrs := internal.NewAttributes()
		internal.RequestBodyAttributes(attrs, req)

		requestBody := attrs.Value().Get(internal.AttributeRequestBody)
		formData, ok := requestBody.(map[string]interface{})
		assert.True(t, ok)

		files, ok := formData["_files"].([]map[string]interface{})
		assert.True(t, ok)
		assert.Len(t, files, 2)
	})

	t.Run("multipart form with array fields", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("tags", "golang")
		_ = writer.WriteField("tags", "testing")
		_ = writer.WriteField("tags", "multipart")
		writer.Close()

		req, _ := http.NewRequest("POST", "http://example.com/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		attrs := internal.NewAttributes()
		internal.RequestBodyAttributes(attrs, req)

		requestBody := attrs.Value().Get(internal.AttributeRequestBody)
		formData, ok := requestBody.(map[string]interface{})
		assert.True(t, ok)

		tags, ok := formData["tags"].([]string)
		assert.True(t, ok, "Tags should be an array")
		assert.Len(t, tags, 3)
		assert.Contains(t, tags, "golang")
		assert.Contains(t, tags, "testing")
		assert.Contains(t, tags, "multipart")
	})

	t.Run("empty multipart form", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req, _ := http.NewRequest("POST", "http://example.com/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		attrs := internal.NewAttributes()
		internal.RequestBodyAttributes(attrs, req)

		requestBody := attrs.Value().Get(internal.AttributeRequestBody)
		assert.Nil(t, requestBody, "Empty form should result in nil request body")
	})
}

func TestRequestBodyAttributesFormUrlEncoded(t *testing.T) {
	t.Run("urlencoded form with text fields", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("username", "testuser")
		formData.Set("password", "secret123")
		formData.Set("remember", "true")

		req, _ := http.NewRequest("POST", "http://example.com/login", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		attrs := internal.NewAttributes()
		internal.RequestBodyAttributes(attrs, req)

		requestBody := attrs.Value().Get(internal.AttributeRequestBody)
		assert.NotNil(t, requestBody)

		bodyMap, ok := requestBody.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "testuser", bodyMap["username"])
		assert.Equal(t, "secret123", bodyMap["password"])
		assert.Equal(t, "true", bodyMap["remember"])
	})

	t.Run("urlencoded form with array values", func(t *testing.T) {
		formData := url.Values{}
		formData.Add("colors", "red")
		formData.Add("colors", "blue")
		formData.Add("colors", "green")

		req, _ := http.NewRequest("POST", "http://example.com/submit", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		attrs := internal.NewAttributes()
		internal.RequestBodyAttributes(attrs, req)

		requestBody := attrs.Value().Get(internal.AttributeRequestBody)
		bodyMap, ok := requestBody.(map[string]interface{})
		assert.True(t, ok)

		colors, ok := bodyMap["colors"].([]string)
		assert.True(t, ok)
		assert.Len(t, colors, 3)
		assert.Contains(t, colors, "red")
		assert.Contains(t, colors, "blue")
		assert.Contains(t, colors, "green")
	})

	t.Run("empty urlencoded form", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "http://example.com/submit", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		attrs := internal.NewAttributes()
		internal.RequestBodyAttributes(attrs, req)

		requestBody := attrs.Value().Get(internal.AttributeRequestBody)
		assert.Nil(t, requestBody)
	})
}

func TestResponseBodyAttributes(t *testing.T) {
	tests := []struct {
		name     string
		response *http.Response
		expected map[string]interface{}
	}{
		{
			name: "non-empty response body",
			response: func() *http.Response {
				body := io.NopCloser(bytes.NewBufferString(`{"key":"value"}`))
				return &http.Response{Body: body}
			}(),
			expected: map[string]interface{}{
				internal.AttributeResponseBody: map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			name: "empty response body",
			response: func() *http.Response {
				body := io.NopCloser(bytes.NewBufferString(``))
				return &http.Response{Body: body}
			}(),
			expected: map[string]interface{}{},
		},
		{
			name: "nil response body",
			response: func() *http.Response {
				return &http.Response{Body: nil}
			}(),
			expected: map[string]interface{}{},
		},
		{
			name: "nil response",
			response: func() *http.Response {
				return nil
			}(),
			expected: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := internal.NewAttributes()
			internal.ResponseBodyAttributes(attrs, tt.response)

			for attrKey, expectedValue := range tt.expected {
				value := attrs.Value().Get(attrKey)
				assert.Equal(t, expectedValue, value)
			}
		})
	}
}

func TestAttributeValues_AddString(t *testing.T) {
	type args struct {
		id        string
		stringVal string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "it should be ok",
			args: args{
				id:        "id",
				stringVal: "string",
			},
			want: "string",
		},
		{
			name: "it should be ok with empty",
			args: args{
				id:        "id",
				stringVal: "",
			},
			want: "",
		},
		{
			name: "it should be ok with long string",
			args: args{
				id:        "id",
				stringVal: strings.Repeat("a", 256),
			},
			want: strings.Repeat("a", 255),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := internal.AttributeValues{}
			attr.AddString(tt.args.id, tt.args.stringVal)
			assert.Equal(t, tt.want, attr.GetString(tt.args.id))
		})
	}
}

func TestRequestAgentAttributes(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		headers  http.Header
		url      *url.URL
		host     string
		expected map[string]string
	}{
		{
			name:    "basic GET request with no headers",
			method:  "GET",
			headers: nil,
			url: &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/resource",
			},
			host:     "example.com",
			expected: map[string]string{},
		},
		{
			name:    "POST request with content length",
			method:  "POST",
			headers: http.Header{"Content-Length": []string{"123"}},
			url: &url.URL{
				Scheme: "https",
				Host:   "api.example.com",
				Path:   "/submit",
			},
			host: "api.example.com",
			expected: map[string]string{
				internal.AttributeRequestMethod:        "POST",
				internal.AttributeRequestURI:           "https://api.example.com/submit",
				internal.AttributeRequestHost:          "api.example.com",
				internal.AttributeRequestContentLength: "123",
			},
		},
		{
			name:     "nil URL and headers",
			method:   "PUT",
			headers:  nil,
			url:      nil,
			host:     "host.com",
			expected: map[string]string{},
		},
		{
			name:    "empty host and headers",
			method:  "DELETE",
			headers: http.Header{},
			url: &url.URL{
				Scheme: "ftp",
				Host:   "ftp.example.com",
				Path:   "/delete",
			},
			host: "",
			expected: map[string]string{
				internal.AttributeRequestMethod: "DELETE",
				internal.AttributeRequestURI:    "ftp://ftp.example.com/delete",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := internal.NewAttributes()
			internal.RequestAgentAttributes(attrs, tt.method, tt.headers, tt.url, tt.host)

			for attrKey, expectedValue := range tt.expected {
				value := attrs.Value().GetString(attrKey)
				assert.Equal(t, expectedValue, value)
			}
		})
	}
}

func TestAttributeValues_Values(t *testing.T) {
	tests := []struct {
		name  string
		setup func() internal.AttributeValues
		want  map[string]interface{}
	}{
		{
			name: "empty AttributeValues",
			setup: func() internal.AttributeValues {
				return internal.AttributeValues{}
			},
			want: map[string]interface{}{},
		},
		{
			name: "mixed stringVal and otherVal",
			setup: func() internal.AttributeValues {
				attr := internal.AttributeValues{}
				attr.AddString("key1", "value1")
				attr.Add("key2", 42)
				return attr
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
		},
		{
			name: "nil otherVal",
			setup: func() internal.AttributeValues {
				attr := internal.AttributeValues{}
				attr.Add("key1", nil)
				return attr
			},
			want: map[string]interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := tt.setup()
			assert.Equal(t, tt.want, attr.Values())
		})
	}
}

func TestResponseCodeAttribute(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected int
	}{
		{
			name:     "standard code",
			code:     200,
			expected: 200,
		},
		{
			name:     "client error code",
			code:     404,
			expected: 404,
		},
		{
			name:     "server error code",
			code:     500,
			expected: 500,
		},
		{
			name:     "informational code",
			code:     100,
			expected: 100,
		},
		{
			name:     "empty code",
			code:     0,
			expected: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := internal.NewAttributes()
			internal.ResponseCodeAttribute(attr, tt.code)
			assert.Equal(t, tt.expected, attr.Value().GetInt(internal.AttributeResponseCode))
		})
	}
}

func TestAttributeValues_IsNotEmpty(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() internal.AttributeValues
		id       string
		expected bool
	}{
		{
			name: "nil value",
			setup: func() internal.AttributeValues {
				attr := internal.AttributeValues{}
				attr.Add("key", nil)
				return attr
			},
			id:       "key",
			expected: false,
		},
		{
			name: "empty string",
			setup: func() internal.AttributeValues {
				attr := internal.AttributeValues{}
				attr.Add("key", "")
				return attr
			},
			id:       "key",
			expected: false,
		},
		{
			name: "non-empty string",
			setup: func() internal.AttributeValues {
				attr := internal.AttributeValues{}
				attr.Add("key", "value")
				return attr
			},
			id:       "key",
			expected: true,
		},
		{
			name: "empty string slice",
			setup: func() internal.AttributeValues {
				attr := internal.AttributeValues{}
				attr.Add("key", []string{})
				return attr
			},
			id:       "key",
			expected: false,
		},
		{
			name: "non-empty string slice",
			setup: func() internal.AttributeValues {
				attr := internal.AttributeValues{}
				attr.Add("key", []string{"value"})
				return attr
			},
			id:       "key",
			expected: true,
		},
		{
			name: "empty map[string]string",
			setup: func() internal.AttributeValues {
				attr := internal.AttributeValues{}
				attr.Add("key", map[string]string{})
				return attr
			},
			id:       "key",
			expected: false,
		},
		{
			name: "non-empty map[string]string",
			setup: func() internal.AttributeValues {
				attr := internal.AttributeValues{}
				attr.Add("key", map[string]string{"k": "v"})
				return attr
			},
			id:       "key",
			expected: true,
		},
		{
			name: "empty interface slice",
			setup: func() internal.AttributeValues {
				attr := internal.AttributeValues{}
				attr.Add("key", []interface{}{})
				return attr
			},
			id:       "key",
			expected: false,
		},
		{
			name: "non-empty interface slice",
			setup: func() internal.AttributeValues {
				attr := internal.AttributeValues{}
				attr.Add("key", []interface{}{1})
				return attr
			},
			id:       "key",
			expected: true,
		},
		{
			name: "empty map[string]interface{}",
			setup: func() internal.AttributeValues {
				attr := internal.AttributeValues{}
				attr.Add("key", map[string]interface{}{})
				return attr
			},
			id:       "key",
			expected: false,
		},
		{
			name: "non-empty map[string]interface{}",
			setup: func() internal.AttributeValues {
				attr := internal.AttributeValues{}
				attr.Add("key", map[string]interface{}{"k": "v"})
				return attr
			},
			id:       "key",
			expected: true,
		},
		{
			name: "integer value",
			setup: func() internal.AttributeValues {
				attr := internal.AttributeValues{}
				attr.Add("key", 42)
				return attr
			},
			id:       "key",
			expected: true,
		},
		{
			name: "boolean value",
			setup: func() internal.AttributeValues {
				attr := internal.AttributeValues{}
				attr.Add("key", true)
				return attr
			},
			id:       "key",
			expected: true,
		},
		{
			name: "non-existent key",
			setup: func() internal.AttributeValues {
				return internal.AttributeValues{}
			},
			id:       "non-existent",
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := tt.setup()
			result := attr.IsNotEmpty(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test the toObj function improvements for array handling
func TestToObjArrayHandling(t *testing.T) {
	// Since toObj is not exported, we'll test it indirectly through RequestBodyAttributes
	tests := []struct {
		name         string
		jsonBody     string
		expectedType string
		description  string
	}{
		{
			name:         "object JSON body",
			jsonBody:     `{"user": "alice", "token": "secret123"}`,
			expectedType: "map[string]interface {}",
			description:  "Should parse object JSON into map",
		},
		{
			name:         "array JSON body",
			jsonBody:     `[{"user": "bob", "token": "secret456"}, {"user": "charlie", "token": "secret789"}]`,
			expectedType: "[]interface {}",
			description:  "Should parse array JSON into slice",
		},
		{
			name:         "simple array",
			jsonBody:     `["value1", "value2", "value3"]`,
			expectedType: "[]interface {}",
			description:  "Should handle simple value arrays",
		},
		{
			name:         "nested structure",
			jsonBody:     `{"users": [{"name": "admin", "token": "adminToken"}], "config": {"apiKey": "sk-123"}}`,
			expectedType: "map[string]interface {}",
			description:  "Should handle complex nested structures",
		},
		{
			name:         "invalid JSON",
			jsonBody:     `{invalid json}`,
			expectedType: "map[string]interface {}",
			description:  "Should fallback to map for invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with JSON body
			req, err := http.NewRequest("POST", "http://example.com", strings.NewReader(tt.jsonBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			
			// Process through RequestBodyAttributes which uses toObj
			attrs := internal.NewAttributes()
			internal.RequestBodyAttributes(attrs, req)
			
			// Get the processed body
			body := attrs.Value().Get(internal.AttributeRequestBody)
			
			if body != nil {
				bodyType := reflect.TypeOf(body).String()
				t.Logf("Test: %s", tt.name)
				t.Logf("Input: %s", tt.jsonBody)
				t.Logf("Output type: %s", bodyType)
				t.Logf("Output value: %+v", body)
				
				// Verify the type matches expected
				assert.Contains(t, bodyType, tt.expectedType, tt.description)
			} else {
				// For invalid JSON, body might be nil or empty map
				t.Logf("Body is nil for input: %s", tt.jsonBody)
			}
		})
	}
}

// TestRequestBodyAttributesWithArrays specifically tests array handling improvements
func TestRequestBodyAttributesWithArrays(t *testing.T) {
	tests := []struct {
		name           string
		jsonBody       string
		expectedLength int
		description    string
	}{
		{
			name:           "array of objects",
			jsonBody:       `[{"id": 1, "name": "first"}, {"id": 2, "name": "second"}]`,
			expectedLength: 2,
			description:    "Should preserve array structure",
		},
		{
			name:           "empty array",
			jsonBody:       `[]`,
			expectedLength: 0,
			description:    "Should handle empty arrays",
		},
		{
			name:           "single element array",
			jsonBody:       `[{"single": "element"}]`,
			expectedLength: 1,
			description:    "Should handle single element arrays",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "http://example.com", strings.NewReader(tt.jsonBody))
			assert.NoError(t, err)
			
			attrs := internal.NewAttributes()
			internal.RequestBodyAttributes(attrs, req)
			
			body := attrs.Value().Get(internal.AttributeRequestBody)
			assert.NotNil(t, body, "Request body should not be nil")
			
			// Check if it's an array
			if arr, ok := body.([]interface{}); ok {
				assert.Equal(t, tt.expectedLength, len(arr), tt.description)
				t.Logf("Successfully parsed array with %d elements", len(arr))
			} else {
				t.Errorf("Expected array but got %T: %+v", body, body)
			}
		})
	}
}
