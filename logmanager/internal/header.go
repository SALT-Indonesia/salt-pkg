package internal

type Header struct {
	Data           map[string]string
	AllowedHeaders []string
	IsDebugMode    bool
}

func ToMapString(headers interface{}) map[string]string {
	if headers == nil {
		return map[string]string{}
	}

	// Safely assert the type
	if hMap, ok := headers.(map[string]string); ok {
		return hMap
	}

	// Return an empty map if the type assertion fails
	return map[string]string{}
}

func (h Header) FilterHeaders() map[string]string {
	if h.Data == nil {
		return map[string]string{}
	}

	if h.IsDebugMode {
		// Return all headers when debug mode is enabled
		return h.Data
	}

	// Otherwise, filter only exposed headers
	result := make(map[string]string)
	for _, exposed := range h.AllowedHeaders {
		if value, ok := h.Data[exposed]; ok {
			result[exposed] = value
		}
	}
	return result
}
