package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const attributeValueLengthLimit = 255

const (
	AttributeResponseBody         = "response"
	AttributeResponseCode         = "status"
	AttributeRequestMethod        = "method"
	AttributeRequestURI           = "url"
	AttributeRequestHost          = "host"
	AttributeRequestContentLength = "request.headers.contentLength"
	AttributeRequestBody          = "request"
	AttributeRequestHeaders       = "headers"
	AttributeResponseHeaders      = "response_headers"
	AttributeDatabaseTable        = "table"
	AttributeDatabaseQuery        = "query"
	AttributeDatabaseHost         = "host"
	AttributeConsumerRequestBody  = "request"
	AttributeConsumerExchange     = "exchange"
	AttributeConsumerQueue        = "queue"
	AttributeConsumerRoutingKey   = "routing_key"
	AttributeRequestQueryParam    = "query_param"
)

type Attributes struct {
	value AttributeValues
}

func (a *Attributes) Value() AttributeValues {
	return a.value
}

type AttributeValue struct {
	stringVal string
	otherVal  interface{}
}

type AttributeValues map[string]AttributeValue

// Values return a map of all id keys with their corresponding attribute values, preferring stringVal over otherVal.
func (attr AttributeValues) Values() map[string]interface{} {
	values := make(map[string]interface{})
	for id, value := range attr {
		if value.stringVal != "" {
			values[id] = value.stringVal
			continue
		}
		values[id] = value.otherVal
	}
	return values
}

// GetString retrieves the string value associated with the given id. Returns an empty string if the id is not found.
func (attr AttributeValues) GetString(id string) string {
	if value, ok := attr[id]; ok {
		return value.stringVal
	}
	return ""
}

func (attr AttributeValues) GetInt(id string) int {
	attrVal := attr.Get(id)
	if attrVal == nil {
		return 0
	}
	return attrVal.(int)
}

// AddString inserts a string value associated with the given id into the AttributeValues map, truncating if necessary.
func (attr AttributeValues) AddString(id string, stringVal string) {
	if stringVal != "" {
		attr[id] = AttributeValue{
			stringVal: truncateStringValueIfLong(stringVal),
		}
	}
}

// Get retrieves the 'otherVal' associated with the given id from the AttributeValues map. Returns nil if not found.
func (attr AttributeValues) Get(id string) interface{} {
	if value, ok := attr[id]; ok {
		return value.otherVal
	}
	return nil
}

// IsEmpty determines whether the value associated with the given id is considered empty according to isEmpty function.
func (attr AttributeValues) IsEmpty(id string) bool {
	return isEmpty(attr.Get(id))
}

// IsNotEmpty checks whether the value associated with the given id is not empty, according to the isEmpty function.
func (attr AttributeValues) IsNotEmpty(id string) bool {
	return !isEmpty(attr.Get(id))
}

func isEmpty(val interface{}) bool {
	switch v := val.(type) {
	case nil:
		return true
	case string:
		return v == ""
	case []string:
		return len(v) == 0
	case map[string]string:
		return len(v) == 0
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	}
	return false
}

// Add inserts an attribute value associated with the given id into the AttributeValues map if otherVal is not nil.
func (attr AttributeValues) Add(id string, otherVal interface{}) {
	if nil != otherVal {
		attr[id] = AttributeValue{
			otherVal: otherVal,
		}
	}
}

// RequestAgentAttributes populates the Attributes object with HTTP request-related data, such as method, URI, and host.
// It registers the request method and checks if the provided URL is not nil to add the request URI.
// If the HTTP header is provided, it also adds the request host and content length to the Attributes.
func RequestAgentAttributes(a *Attributes, method string, h http.Header, u *url.URL, host string) {
	a.value.AddString(AttributeRequestMethod, method)

	if nil != u {
		a.value.AddString(AttributeRequestURI, safeURL(u))
		a.value.Add(AttributeRequestQueryParam, QueryParamsToMap(u))
	}

	if nil == h {
		return
	}
	a.value.AddString(AttributeRequestHost, host)

	if l := getContentLengthFromHeader(h); l >= 0 {
		a.value.AddString(AttributeRequestContentLength, toString(l))
	}

	headerAttributes(a, h)
}

// QueryParamsToMap extracts query parameters from a URL and returns them as a map with keys and their first values.
func QueryParamsToMap(u *url.URL) map[string]string {
	result := make(map[string]string)
	for k, v := range u.Query() {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

func headerAttributes(a *Attributes, h http.Header) {
	headers := make(map[string]string)
	for k, v := range h {
		headers[k] = strings.Join(v, ",")
	}
	a.value.Add(AttributeRequestHeaders, headers)
}

// RequestBodyAttributes reads the body from an http.Request and adds it as an attribute to the provided Attributes object.
// If the request or its body is nil, no action is taken. The request body is converted to a suitable format and stored under
// the "request.body" key in the Attributes map. This function ensures the request's body can be read multiple times by
// resetting the body after reading it.
// For multipart/form-data requests, it parses the form and extracts all fields including file information.
func RequestBodyAttributes(a *Attributes, r *http.Request) {
	if nil == r {
		return
	}

	contentType := r.Header.Get("Content-Type")

	// Handle multipart/form-data
	if strings.Contains(contentType, "multipart/form-data") {
		parseMultipartFormData(a, r)
		headerAttributes(a, r.Header)
		return
	}

	// Handle application/x-www-form-urlencoded
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		parseFormData(a, r)
		headerAttributes(a, r.Header)
		return
	}

	// Handle JSON and other content types
	if nil == r.Body {
		return
	}

	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if len(bodyBytes) > 0 {
		a.value.Add(AttributeRequestBody, toObj(bodyBytes))
	}

	headerAttributes(a, r.Header)
}

// parseMultipartFormData parses multipart/form-data and extracts form fields and file information.
func parseMultipartFormData(a *Attributes, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return
	}

	formData := make(map[string]interface{})

	// Extract regular form fields
	if r.MultipartForm != nil && r.MultipartForm.Value != nil {
		for key, values := range r.MultipartForm.Value {
			if len(values) == 1 {
				formData[key] = values[0]
			} else if len(values) > 1 {
				formData[key] = values
			}
		}
	}

	// Extract file information
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		files := make([]map[string]interface{}, 0)
		for fieldName, fileHeaders := range r.MultipartForm.File {
			for _, fileHeader := range fileHeaders {
				fileInfo := map[string]interface{}{
					"field":    fieldName,
					"filename": fileHeader.Filename,
					"size":     fileHeader.Size,
					"header":   fileHeader.Header,
				}
				files = append(files, fileInfo)
			}
		}
		if len(files) > 0 {
			formData["_files"] = files
		}
	}

	if len(formData) > 0 {
		a.value.Add(AttributeRequestBody, formData)
	}
}

// parseFormData parses application/x-www-form-urlencoded form data.
func parseFormData(a *Attributes, r *http.Request) {
	// Read body first and restore it after parsing
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := r.ParseForm(); err != nil {
		return
	}

	// Restore body again after ParseForm consumed it
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	formData := make(map[string]interface{})
	for key, values := range r.Form {
		if len(values) == 1 {
			formData[key] = values[0]
		} else if len(values) > 1 {
			formData[key] = values
		}
	}

	if len(formData) > 0 {
		a.value.Add(AttributeRequestBody, formData)
	}
}

// RequestBodyAttribute adds a request body to the given attributes if the body is not nil.
func RequestBodyAttribute(a *Attributes, body interface{}) {
	if nil == body {
		return
	}
	a.value.Add(AttributeRequestBody, body)
}

// ResponseBodyAttributes updates the Attributes with the response body content if present.
// It reads all bytes from the response body and resets the body to allow further reading.
// The content is added to Attributes if it is not empty.
func ResponseBodyAttributes(a *Attributes, resp *http.Response) {
	if nil == resp {
		return
	}

	if nil == resp.Body {
		return
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if len(bodyBytes) > 0 {
		a.value.Add(AttributeResponseBody, toObj(bodyBytes))
	}
}

func toString(val interface{}) string {
	return fmt.Sprintf("%v", val)
}

// ResponseCodeAttribute assigns an HTTP status code as a string to the given Attributes instance using the AttributeResponseCode key.
func ResponseCodeAttribute(a *Attributes, code int) {
	a.value.Add(AttributeResponseCode, code)
}

func toObj(bodyBytes []byte) interface{} {
	// First try to unmarshal as a generic interface{} to handle both objects and arrays
	var payload interface{}
	err := json.Unmarshal(bodyBytes, &payload)
	if err != nil {
		// If JSON unmarshaling fails, try as map[string]interface{} for backward compatibility
		var mapPayload map[string]interface{}
		_ = json.Unmarshal(bodyBytes, &mapPayload)
		return mapPayload
	}
	return payload
}

func truncateStringValueIfLong(val string) string {
	if len(val) > attributeValueLengthLimit {
		return stringLengthByteLimit(val, attributeValueLengthLimit)
	}
	return val
}

func safeURL(u *url.URL) string {
	if nil == u {
		return ""
	}
	if "" != u.Opaque {
		return ""
	}

	ur := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   u.Path,
	}
	return ur.String()
}

func stringLengthByteLimit(str string, byteLimit int) string {
	if len(str) <= byteLimit {
		return str
	}

	limitIndex := 0
	for pos := range str {
		if pos > byteLimit {
			break
		}
		limitIndex = pos
	}
	return str[0:limitIndex]
}

func getContentLengthFromHeader(h http.Header) int64 {
	if cl := h.Get("Content-Length"); cl != "" {
		if contentLength, err := strconv.ParseInt(cl, 10, 64); err == nil {
			return contentLength
		}
	}

	return -1
}

func NewAttributes() *Attributes {
	return &Attributes{
		value: make(AttributeValues),
	}
}
