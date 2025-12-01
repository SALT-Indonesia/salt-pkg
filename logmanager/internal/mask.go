package internal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/ggwhite/go-masker/v2"
	"github.com/oliveagle/jsonpath"
	"github.com/sirupsen/logrus"
)

// MaskType defines the masking strategy
type MaskType int

const (
	// FullMask completely replaces the value with asterisks
	FullMask MaskType = iota

	// PartialMask shows first and last few characters, rest are masked
	PartialMask

	// HideMask hides the value entirely without displaying any characters.
	HideMask

	// EmailMask masks email addresses preserving domain and showing first/last chars of username
	// Example: arfan.azhari@salt.id → ar******ri@salt.id
	EmailMask
)

// MaskingConfig defines how a specific field should be masked using JSONPath or field pattern
type MaskingConfig struct {
	JSONPath     string   // JSONPath expression to identify fields to mask
	FieldPattern string   // Field name pattern to match (e.g., "password" will match any field containing "password")
	Field        string   // Exact field name to match (for backward compatibility)
	Type         MaskType // Type of masking to apply
	ShowFirst    int      // Number of characters to show from the start (for PartialMask)
	ShowLast     int      // Number of characters to show from the end (for PartialMask)
}

// JSONMasker provides advanced masking functionality
type JSONMasker struct {
	maskingConfigs   []MaskingConfig
	enableStructTags bool
	structTagMasker  *masker.MaskerMarshaler
}

// NewJSONMasker creates a new masker with specified MaskingConfig configurations
func NewJSONMasker(maskingConfigs []MaskingConfig) *JSONMasker {
	return &JSONMasker{
		maskingConfigs:   maskingConfigs,
		enableStructTags: false,
		structTagMasker:  masker.NewMaskerMarshaler(),
	}
}

// NewJSONMaskerWithStructTags creates a new masker with struct tag support enabled
func NewJSONMaskerWithStructTags(maskingConfigs []MaskingConfig) *JSONMasker {
	return &JSONMasker{
		maskingConfigs:   maskingConfigs,
		enableStructTags: true,
		structTagMasker:  masker.NewMaskerMarshaler(),
	}
}

// MaskData applies masking to the provided data and returns the masked result
// This is a public method primarily for testing purposes
func (m *JSONMasker) MaskData(data interface{}) interface{} {
	// First apply struct tag masking if enabled
	if m.enableStructTags && m.shouldApplyStructTagMasking(data) {
		if maskedData, err := m.structTagMasker.Struct(data); err == nil {
			data = maskedData
		}
	}

	// Then apply existing field/JSONPath masking
	return m.maskFields(data)
}

// getAllMaskingConfigs returns all masking configurations
func (m *JSONMasker) getAllMaskingConfigs() []MaskingConfig {
	return m.maskingConfigs
}

// shouldApplyStructTagMasking determines if struct tag masking should be applied
// It checks if the data is a struct or a pointer to struct that might have mask tags
func (m *JSONMasker) shouldApplyStructTagMasking(data interface{}) bool {
	if data == nil {
		return false
	}

	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	// Handle pointers
	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		_ = v.Elem() // Dereference pointer
		t = t.Elem()
	}

	// Only apply to structs
	if t.Kind() != reflect.Struct {
		return false
	}

	// Check if any field has a mask tag
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("mask") != "" {
			return true
		}
	}

	return false
}

// maskValueWithMaskingConfig applies the appropriate masking based on MaskingConfig
func (m *JSONMasker) maskValueWithMaskingConfig(value interface{}, config MaskingConfig) interface{} {
	// Convert value to string for masking
	strValue := fmt.Sprintf("%v", value)

	// Handle empty or very short values
	if len(strValue) <= 1 {
		return strings.Repeat("*", len(strValue))
	}

	// Handle too long characters
	if len(strValue) >= 255 {
		return "*"
	}

	// Handle different mask types
	switch config.Type {
	case HideMask:
		return "*"

	case FullMask:
		return strings.Repeat("*", len(strValue))

	case PartialMask:
		// Ensure we don't show more characters than the string length
		showFirst := min(config.ShowFirst, len(strValue))
		showLast := min(config.ShowLast, len(strValue))

		// If showing first and last covers most of the string, fall back to full mask
		if showFirst+showLast >= len(strValue) {
			return strings.Repeat("*", len(strValue))
		}

		return strValue[:showFirst] +
			strings.Repeat("*", len(strValue)-showFirst-showLast) +
			strValue[len(strValue)-showLast:]

	case EmailMask:
		return m.maskEmail(strValue, config)

	default:
		return value
	}
}

// maskEmail masks an email address preserving the domain and showing first/last chars of username
// Example: arfan.azhari@salt.id → ar******ri@salt.id
func (m *JSONMasker) maskEmail(email string, config MaskingConfig) string {
	// Find @ position
	atIndex := strings.LastIndex(email, "@")
	if atIndex == -1 {
		// Not a valid email, fall back to partial mask
		return m.maskValueWithMaskingConfig(email, MaskingConfig{
			Type:      PartialMask,
			ShowFirst: config.ShowFirst,
			ShowLast:  config.ShowLast,
		}).(string)
	}

	username := email[:atIndex]
	domain := email[atIndex:] // includes @

	// Use config values or defaults (2 chars first, 2 chars last)
	showFirst := config.ShowFirst
	if showFirst == 0 {
		showFirst = 2
	}
	showLast := config.ShowLast
	if showLast == 0 {
		showLast = 2
	}

	// Handle short usernames
	if len(username) <= showFirst+showLast {
		// Username too short to mask meaningfully, show first char and mask rest
		if len(username) <= 1 {
			return "*" + domain
		}
		return string(username[0]) + strings.Repeat("*", len(username)-1) + domain
	}

	// Mask username: show first N and last N chars, mask the middle
	maskedUsername := username[:showFirst] +
		strings.Repeat("*", len(username)-showFirst-showLast) +
		username[len(username)-showLast:]

	return maskedUsername + domain
}

// maskFields recursively masks fields in a nested structure
func (m *JSONMasker) maskFields(data interface{}) interface{} {
	// First, handle JSONPath configurations that operate on the entire data structure
	data = m.applyJSONPathMasking(data)

	// Then handle field-level masking recursively
	// Handle map[string]interface{}
	if mv, ok := data.(map[string]interface{}); ok {
		masked := make(map[string]interface{})
		for k, v := range mv {
			// Check if this field needs masking
			masked[k] = m.maskFieldByConfig(k, v)
		}
		return masked
	}

	// Handle map[string]string
	if mv, ok := data.(map[string]string); ok {
		masked := make(map[string]interface{})
		for k, v := range mv {
			// Check if this field needs masking
			masked[k] = m.maskFieldByConfig(k, v)
		}
		return masked
	}

	// Handle slice/array
	if sv, ok := data.([]interface{}); ok {
		masked := make([]interface{}, len(sv))
		for i, item := range sv {
			masked[i] = m.maskFields(item)
		}
		return masked
	}

	return data
}

// applyJSONPathMasking applies JSONPath-based masking to the entire data structure
func (m *JSONMasker) applyJSONPathMasking(data interface{}) interface{} {
	// Get all masking configurations that use JSONPath
	jsonPathConfigs := make([]MaskingConfig, 0)
	for _, config := range m.maskingConfigs {
		if config.JSONPath != "" {
			jsonPathConfigs = append(jsonPathConfigs, config)
		}
	}

	// If no JSONPath configs, return data unchanged
	if len(jsonPathConfigs) == 0 {
		return data
	}

	// Convert data to map[string]interface{} for JSONPath processing
	dataMap, err := m.convertToMap(data)
	if err != nil {
		return data // Return original data if conversion fails
	}

	// Apply each JSONPath configuration
	for _, config := range jsonPathConfigs {
		dataMap = m.applyJSONPathConfig(dataMap, config)
	}

	return dataMap
}

// convertToMap converts various data types to interface{} for JSONPath processing
func (m *JSONMasker) convertToMap(data interface{}) (interface{}, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		return v, nil
	case []interface{}:
		return v, nil
	case map[string]string:
		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = val
		}
		return result, nil
	default:
		// Try JSON marshal/unmarshal for other types
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		var result interface{}
		err = json.Unmarshal(jsonBytes, &result)
		if err != nil {
			return nil, err
		}

		return result, nil
	}
}

// applyJSONPathConfig applies a single JSONPath configuration to the data
func (m *JSONMasker) applyJSONPathConfig(data interface{}, config MaskingConfig) interface{} {
	// Check if this is a recursive descent pattern ($..fieldname)
	if strings.HasPrefix(config.JSONPath, "$..") {
		fieldName := strings.TrimPrefix(config.JSONPath, "$..")
		return m.recursiveMaskField(data, fieldName, config)
	}

	// For array at root level with $[*] pattern, handle specially
	if strings.HasPrefix(config.JSONPath, "$[*]") {
		if arr, ok := data.([]interface{}); ok {
			fieldPath := strings.TrimPrefix(config.JSONPath, "$[*].")
			if fieldPath == config.JSONPath {
				// No field specified after $[*], return as-is
				return data
			}
			// Mask the specified field in each array element
			result := make([]interface{}, len(arr))
			for i, item := range arr {
				if itemMap, ok := item.(map[string]interface{}); ok {
					maskedItem := make(map[string]interface{})
					for k, v := range itemMap {
						if k == fieldPath {
							maskedItem[k] = m.maskValueWithMaskingConfig(v, config)
						} else {
							maskedItem[k] = v
						}
					}
					result[i] = maskedItem
				} else {
					result[i] = item
				}
			}
			return result
		}
	}

	// Convert to appropriate type for JSONPath processing
	// This now handles both maps and arrays
	processableData := data
	
	// Use JSONPath to find matching values
	result, err := jsonpath.JsonPathLookup(processableData, config.JSONPath)
	if err != nil {
		// If JSONPath doesn't match anything, return data unchanged
		return data
	}

	// Apply masking based on the result type
	switch values := result.(type) {
	case []interface{}:
		// Multiple matches - mask each one
		for _, value := range values {
			data = m.maskJSONPathValue(data, config.JSONPath, value, config)
		}
	default:
		// Single match - mask it
		data = m.maskJSONPathValue(data, config.JSONPath, result, config)
	}

	return data
}

// recursiveMaskField recursively masks all occurrences of a field name at any depth
func (m *JSONMasker) recursiveMaskField(data interface{}, fieldName string, config MaskingConfig) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			// Case-insensitive contains check for field name
			if strings.Contains(strings.ToLower(k), strings.ToLower(fieldName)) {
				// Mask this field
				result[k] = m.maskValueWithMaskingConfig(val, config)
			} else {
				// Recursively process nested structures
				result[k] = m.recursiveMaskField(val, fieldName, config)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = m.recursiveMaskField(item, fieldName, config)
		}
		return result
	default:
		// For primitive types, return as-is
		return v
	}
}

// maskJSONPathValue masks a specific value found by JSONPath
func (m *JSONMasker) maskJSONPathValue(data interface{}, _ string, originalValue interface{}, config MaskingConfig) interface{} {
	maskedValue := m.maskValueWithMaskingConfig(originalValue, config)

	// Replace the original value with the masked value in the data structure
	// This is a simplified approach - in a production system, you might want more sophisticated path manipulation
	return m.replaceValueAtPath(data, originalValue, maskedValue)
}

// replaceValueAtPath replaces a value at the specified JSONPath with a masked value
func (m *JSONMasker) replaceValueAtPath(data interface{}, originalValue, maskedValue interface{}) interface{} {
	// Use a recursive approach to find and replace the value
	// This is a simplified implementation
	result := m.recursiveReplace(data, originalValue, maskedValue)
	return result
}

// recursiveReplace recursively searches for the original value and replaces it with the masked value
func (m *JSONMasker) recursiveReplace(data interface{}, originalValue, maskedValue interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			// Use a safer comparison that handles different types
			if m.valuesEqual(val, originalValue) {
				result[k] = maskedValue
			} else {
				result[k] = m.recursiveReplace(val, originalValue, maskedValue)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			if m.valuesEqual(item, originalValue) {
				result[i] = maskedValue
			} else {
				result[i] = m.recursiveReplace(item, originalValue, maskedValue)
			}
		}
		return result
	default:
		if m.valuesEqual(v, originalValue) {
			return maskedValue
		}
		return v
	}
}

// valuesEqual safely compares two values for equality, handling maps and slices
func (m *JSONMasker) valuesEqual(a, b interface{}) bool {
	// For simple types, direct comparison works
	switch a.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return a == b
	case nil:
		return b == nil
	default:
		// For complex types like maps and slices, use JSON comparison as a fallback
		aJSON, errA := json.Marshal(a)
		bJSON, errB := json.Marshal(b)
		if errA != nil || errB != nil {
			return false
		}
		return string(aJSON) == string(bJSON)
	}
}

// maskFieldByConfig masks a specific field based on its configuration
func (m *JSONMasker) maskFieldByConfig(field string, value interface{}) interface{} {
	// Get all masking configurations in a unified format
	allConfigs := m.getAllMaskingConfigs()

	// Check if a field matches any configuration
	for _, config := range allConfigs {
		if m.fieldMatches(field, config) {
			return m.maskValueWithMaskingConfig(value, config)
		}
	}

	// If no specific mask, but value is a complex type, recursively mask
	switch v := value.(type) {
	case map[string]interface{}:
		return m.maskFields(v)
	case map[string]string:
		return m.maskFields(v)
	case []interface{}:
		return m.maskFields(v)
	default:
		return value
	}
}

// fieldMatches checks if a field matches the given configuration
// Note: JSONPath matching is handled separately in applyJSONPathMasking
func (m *JSONMasker) fieldMatches(field string, config MaskingConfig) bool {
	// Check Field first (for exact field matching)
	if config.Field != "" {
		return field == config.Field
	}

	// Check FieldPattern (case-insensitive substring match)
	if config.FieldPattern != "" {
		return strings.Contains(strings.ToLower(field), strings.ToLower(config.FieldPattern))
	}

	// JSONPath configurations are handled in applyJSONPathMasking, not here
	// This method only handles field-level matching
	return false
}

// LogrusMiddleware creates a hook for Logrus to mask JSON fields
func (m *JSONMasker) LogrusMiddleware() logrus.Hook {
	return &jsonMaskingHook{masker: m}
}

// jsonMaskingHook implements logrus.Hook
type jsonMaskingHook struct {
	masker *JSONMasker
}

// Levels define which log levels this hook will be triggered on
func (hook *jsonMaskingHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire masks sensitive fields before logging
func (hook *jsonMaskingHook) Fire(entry *logrus.Entry) error {
	// Create a copy of the entry's data
	data := make(logrus.Fields)

	// Mask each field
	for k, v := range entry.Data {
		// Recursively mask fields
		data[k] = hook.masker.maskFields(v)
	}

	// Replace the entry's data with masked data
	entry.Data = data
	return nil
}
