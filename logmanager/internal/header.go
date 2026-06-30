package internal

import "strings"

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

	// Otherwise, filter only exposed headers. An entry may be either an exact
	// header name or a trailing-"*" wildcard prefix (e.g. "CF-*" exposes every
	// Cloudflare header). Wildcard matching is case-insensitive because Go
	// canonicalizes header keys (e.g. "CF-Connecting-IP" -> "Cf-Connecting-Ip").
	// A bare "*" therefore exposes all headers.
	result := make(map[string]string)
	for _, exposed := range h.AllowedHeaders {
		if prefix, ok := strings.CutSuffix(exposed, "*"); ok {
			lowerPrefix := strings.ToLower(prefix)
			for key, value := range h.Data {
				if strings.HasPrefix(strings.ToLower(key), lowerPrefix) {
					result[key] = value
				}
			}
			continue
		}
		if value, ok := h.Data[exposed]; ok {
			result[exposed] = value
		}
	}
	return result
}
