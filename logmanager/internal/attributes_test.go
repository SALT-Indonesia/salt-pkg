package internal_test

import (
	"bytes"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/url"
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
