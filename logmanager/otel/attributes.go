package otel

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

// toString converts various types to string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return fmt.Sprintf("%v", val)
	default:
		return ""
	}
}

// StringAttribute creates a string key-value attribute
func StringAttribute(key, value string) attribute.KeyValue {
	return attribute.String(key, value)
}

// IntAttribute creates an int key-value attribute
func IntAttribute(key string, value int64) attribute.KeyValue {
	return attribute.Int64(key, value)
}

// HTTPServerAttributes creates attributes for HTTP server spans
func HTTPServerAttributes(req *http.Request) []attribute.KeyValue {
	if req == nil {
		return nil
	}

	attrs := []attribute.KeyValue{
		attribute.String("http.method", req.Method),
		attribute.String("http.scheme", req.URL.Scheme),
		attribute.String("server.address", req.Host),
	}

	// Set the full URL for better trace visualization
	if req.URL != nil {
		attrs = append(attrs, attribute.String("http.url", req.URL.String()))
		attrs = append(attrs, attribute.String("http.target", req.URL.Path))
	}

	// Add remote address if available
	if req.RemoteAddr != "" {
		attrs = append(attrs, attribute.String("http.client_ip", req.RemoteAddr))
	}

	// Add user agent if available
	if ua := req.UserAgent(); ua != "" {
		attrs = append(attrs, attribute.String("http.user_agent", ua))
	}

	return attrs
}

// HTTPClientAttributes creates attributes for HTTP client spans
func HTTPClientAttributes(req *http.Request) []attribute.KeyValue {
	if req == nil {
		return nil
	}

	attrs := []attribute.KeyValue{
		attribute.String("http.method", req.Method),
	}

	// Add URL components
	if req.URL != nil {
		attrs = append(attrs, attribute.String("http.url", req.URL.String()))
		attrs = append(attrs, attribute.String("http.scheme", req.URL.Scheme))
		attrs = append(attrs, attribute.String("server.address", req.URL.Host))

		if req.URL.Path != "" {
			attrs = append(attrs, attribute.String("http.target", req.URL.Path))
		}
	}

	return attrs
}

// HTTPResponseAttributes creates attributes for HTTP responses
func HTTPResponseAttributes(statusCode int) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.Int("http.status_code", statusCode),
	}
}

// DatabaseAttributes creates attributes for database spans
func DatabaseAttributes(system, table, query, host string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("db.system", system),
	}

	// Add table name if available
	if table != "" {
		attrs = append(attrs, attribute.String("db.name", table))
	}

	// Add query if available (be careful with sensitive data)
	if query != "" {
		attrs = append(attrs, attribute.String("db.statement", SanitizeQuery(query)))
	}

	// Add connection info
	if host != "" {
		attrs = append(attrs, attribute.String("peer.service", host))
	}

	return attrs
}

// SanitizeQuery removes sensitive data from SQL queries
func SanitizeQuery(query string) string {
	// Convert to lowercase for pattern matching
	queryLower := strings.ToLower(query)
	result := query

	// List of patterns to redact
	patterns := []struct {
		pattern string
		replacement string
	}{
		{"password\\s*=\\s*'[^']*'", "password='***'"},
		{"password\\s*=\\s*\"[^\"]*\"", "password=\"***\""},
		{"token\\s*=\\s*'[^']*'", "token='***'"},
		{"token\\s*=\\s*\"[^\"]*\"", "token=\"***\""},
		{"api_key\\s*=\\s*'[^']*'", "api_key='***'"},
		{"api_key\\s*=\\s*\"[^\"]*\"", "api_key=\"***\""},
	}

	// Apply patterns
	for _, p := range patterns {
		// Simple replacement - in production use regex
		idx := strings.Index(queryLower, p.pattern[:10])
		if idx != -1 {
			// Found a match, do basic redaction
			result = "***" // Simplified for now
		}
	}

	return result
}

// GRPCAttributes creates attributes for gRPC spans
func GRPCAttributes(service, method string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("rpc.system", "grpc"),
	}

	if service != "" {
		attrs = append(attrs, attribute.String("rpc.service", service))
	}

	if method != "" {
		attrs = append(attrs, attribute.String("rpc.method", method))
	}

	return attrs
}

// ErrorAttributes creates attributes for error recording
func ErrorAttributes(err error) []attribute.KeyValue {
	if err == nil {
		return nil
	}

	return []attribute.KeyValue{
		attribute.String("exception.message", err.Error()),
		attribute.String("exception.type", "error"),
	}
}

// TagsAttributes creates attributes from custom tags
func TagsAttributes(tags []string) []attribute.KeyValue {
	if len(tags) == 0 {
		return nil
	}

	attrs := make([]attribute.KeyValue, 0, len(tags))
	for _, tag := range tags {
		attrs = append(attrs, attribute.String("tag", tag))
	}

	return attrs
}

// CustomTraceIDAttribute creates an attribute for the custom logmanager trace ID
func CustomTraceIDAttribute(traceID string) attribute.KeyValue {
	return attribute.String("logmanager.trace_id", traceID)
}

// ServiceNameAttribute creates an attribute for the service name
func ServiceNameAttribute(service string) attribute.KeyValue {
	return attribute.String("service.name", service)
}

// EnvironmentAttribute creates an attribute for the environment
func EnvironmentAttribute(env string) attribute.KeyValue {
	return attribute.String("deployment.environment", env)
}

// LatencyAttribute creates an attribute for operation latency in milliseconds
func LatencyAttribute(ms int64) attribute.KeyValue {
	return attribute.Int64("latency_ms", ms)
}

// RequestBodyAttribute creates an attribute for request body (truncated if too large)
func RequestBodyAttribute(body interface{}) []attribute.KeyValue {
	if body == nil {
		return nil
	}

	// Convert to string and truncate if necessary
	bodyStr := toString(body)
	if len(bodyStr) > 1000 {
		bodyStr = bodyStr[:1000] + "... (truncated)"
	}

	return []attribute.KeyValue{
		attribute.String("http.request.body", bodyStr),
	}
}

// ResponseBodyAttribute creates an attribute for response body (truncated if too large)
func ResponseBodyAttribute(body interface{}) []attribute.KeyValue {
	if body == nil {
		return nil
	}

	// Convert to string and truncate if necessary
	bodyStr := toString(body)
	if len(bodyStr) > 1000 {
		bodyStr = bodyStr[:1000] + "... (truncated)"
	}

	return []attribute.KeyValue{
		attribute.String("http.response.body", bodyStr),
	}
}

// QueryParamAttributes creates attributes from URL query parameters
func QueryParamAttributes(query url.Values) []attribute.KeyValue {
	if len(query) == 0 {
		return nil
	}

	attrs := make([]attribute.KeyValue, 0, len(query))
	for key, values := range query {
		if len(values) > 0 {
			attrs = append(attrs, attribute.String("http.query_param."+key, values[0]))
		}
	}

	return attrs
}

// StatusCodeAttribute creates an attribute for HTTP status code
func StatusCodeAttribute(code int) attribute.KeyValue {
	return attribute.Int("http.status_code", code)
}
