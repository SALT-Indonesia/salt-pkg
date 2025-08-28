package customv2

type ProcessOrderResponse struct {
	OrderID       string `json:"order_id"`
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id"`
	Message       string `json:"message"`
}

// Custom error response structures for different error scenarios
type ValidationErrorResponse struct {
	Type      string            `json:"type"`
	Message   string            `json:"message"`
	Field     string            `json:"field"`
	Value     interface{}       `json:"value"`
	Code      string            `json:"code"`
	RequestID string            `json:"request_id"`
	Details   map[string]string `json:"details,omitempty"`
}

type BusinessErrorResponse struct {
	Type        string                 `json:"type"`
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Reason      string                 `json:"reason"`
	Timestamp   string                 `json:"timestamp"`
	RequestID   string                 `json:"request_id"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type SystemErrorResponse struct {
	Type      string `json:"type"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
	Service   string `json:"service"`
	TraceID   string `json:"trace_id,omitempty"`
}