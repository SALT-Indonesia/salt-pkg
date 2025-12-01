package logmanager

import (
	"fmt"
	"reflect"

	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"github.com/ggwhite/go-masker/v2"
)

// MaskType defines the masking strategy
type MaskType = internal.MaskType

const (
	// FullMask completely replaces the value with asterisks
	FullMask = internal.FullMask

	// PartialMask shows first and last few characters, rest are masked
	PartialMask = internal.PartialMask

	// HideMask hides the value entirely without displaying any characters.
	HideMask = internal.HideMask

	// EmailMask masks email addresses preserving domain and showing first/last chars of username
	// Example: arfan.azhari@salt.id → ar******ri@salt.id
	EmailMask = internal.EmailMask
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

// ConvertMaskingConfigs converts MaskingConfig slice to internal representation
func ConvertMaskingConfigs(configs []MaskingConfig) []internal.MaskingConfig {
	internalConfigs := make([]internal.MaskingConfig, len(configs))

	for i, config := range configs {
		internalConfigs[i] = internal.MaskingConfig{
			JSONPath:     config.JSONPath,
			FieldPattern: config.FieldPattern,
			Field:        config.Field,
			Type:         config.Type, // No conversion needed anymore
			ShowFirst:    config.ShowFirst,
			ShowLast:     config.ShowLast,
		}
	}

	return internalConfigs
}

// StructMask applies masking to struct fields using struct tags
// This function processes structs with `mask` tags and returns the masked version
// It integrates with the existing logmanager masking system
func StructMask(data interface{}) (interface{}, error) {
	if data == nil {
		return nil, fmt.Errorf("input data cannot be nil")
	}

	// Check if the input is a struct or pointer to struct
	if !isStructOrStructPointer(data) {
		return data, fmt.Errorf("input must be a struct or pointer to struct")
	}

	m := masker.NewMaskerMarshaler()
	return m.Struct(data)
}

// there isStructOrStructPointer checks if the input is a struct or pointer to struct
func isStructOrStructPointer(data interface{}) bool {
	t := reflect.TypeOf(data)
	if t == nil {
		return false
	}

	// Handle pointer
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.Kind() == reflect.Struct
}

// StructMaskWithConfig applies masking to struct fields using both struct tags and additional configurations
// This allows combining go-masker struct tags with logmanager's advanced masking configurations
func StructMaskWithConfig(data interface{}, configs []MaskingConfig) (interface{}, error) {
	// First apply struct tag masking using go-masker (only for structs)
	maskedData, err := StructMask(data)
	if err != nil {
		// If struct masking fails (e.g., input is a map), use original data
		// but still apply JSONPath/FieldPattern masking below
		maskedData = data
	}

	// Then apply additional field/JSONPath configurations using logmanager's masking
	if len(configs) > 0 {
		internalConfigs := ConvertMaskingConfigs(configs)
		jsonMasker := internal.NewJSONMasker(internalConfigs)
		maskedData = jsonMasker.MaskData(maskedData)
	}

	return maskedData, nil
}

// NewJSONMaskerWithStructTags creates a new masker with struct tag support enabled
// This integrates go-masker struct tags with logmanager's advanced masking system
func NewJSONMaskerWithStructTags(configs []MaskingConfig) *internal.JSONMasker {
	internalConfigs := ConvertMaskingConfigs(configs)
	return internal.NewJSONMaskerWithStructTags(internalConfigs)
}
