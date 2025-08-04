package internal_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractHeader(t *testing.T) {
	tests := []struct {
		name           string
		header         internal.Header
		expectedResult map[string]string
	}{
		{
			name: "nil data",
			header: internal.Header{
				Data:           nil,
				AllowedHeaders: []string{"Content-Type", "Authorization"},
				IsDebugMode:    false,
			},
			expectedResult: map[string]string{},
		},
		{
			name: "empty data",
			header: internal.Header{
				Data:           map[string]string{},
				AllowedHeaders: []string{"Content-Type", "Authorization"},
				IsDebugMode:    false,
			},
			expectedResult: map[string]string{},
		},
		{
			name: "no matching exposed headers",
			header: internal.Header{
				Data:           map[string]string{"User-Agent": "Go-http-client", "Host": "example.com"},
				AllowedHeaders: []string{"Content-Type", "Authorization"},
				IsDebugMode:    false,
			},
			expectedResult: map[string]string{},
		},
		{
			name: "matching exposed headers",
			header: internal.Header{
				Data:           map[string]string{"Content-Type": "application/json", "Authorization": "Bearer token"},
				AllowedHeaders: []string{"Content-Type", "Authorization"},
				IsDebugMode:    false,
			},
			expectedResult: map[string]string{"Content-Type": "application/json", "Authorization": "Bearer token"},
		},
		{
			name: "partial matches",
			header: internal.Header{
				Data:           map[string]string{"Content-Type": "application/json", "User-Agent": "Go-http-client"},
				AllowedHeaders: []string{"Content-Type", "Authorization"},
				IsDebugMode:    false,
			},
			expectedResult: map[string]string{"Content-Type": "application/json"},
		},
		{
			name: "debug mode all headers exposed",
			header: internal.Header{
				Data:           map[string]string{"Content-Type": "application/json", "Authorization": "Bearer token", "User-Agent": "Go-http-client"},
				AllowedHeaders: []string{"Content-Type"},
				IsDebugMode:    true,
			},
			expectedResult: map[string]string{"Content-Type": "application/json", "Authorization": "Bearer token", "User-Agent": "Go-http-client"},
		},
		{
			name: "empty exposed headers with debug mode",
			header: internal.Header{
				Data:           map[string]string{"Content-Type": "application/json", "Authorization": "Bearer token"},
				AllowedHeaders: []string{},
				IsDebugMode:    true,
			},
			expectedResult: map[string]string{"Content-Type": "application/json", "Authorization": "Bearer token"},
		},
		{
			name: "empty exposed headers without debug mode",
			header: internal.Header{
				Data:           map[string]string{"Content-Type": "application/json", "Authorization": "Bearer token"},
				AllowedHeaders: []string{},
				IsDebugMode:    false,
			},
			expectedResult: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.header.FilterHeaders()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestToMapString(t *testing.T) {
	tests := []struct {
		name           string
		input          interface{}
		expectedResult map[string]string
	}{
		{
			name:           "nil input",
			input:          nil,
			expectedResult: map[string]string{},
		},
		{
			name:           "valid map input",
			input:          map[string]string{"key1": "value1", "key2": "value2"},
			expectedResult: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:           "empty map input",
			input:          map[string]string{},
			expectedResult: map[string]string{},
		},
		{
			name:           "invalid type input with array integer",
			input:          []int{1, 2, 3},
			expectedResult: map[string]string{},
		},
		{
			name:           "invalid type input with array string",
			input:          []string{"a", "b", "c"},
			expectedResult: map[string]string{},
		},
		{
			name:           "invalid type input with any string or integer",
			input:          []any{"a", "b", 3},
			expectedResult: map[string]string{},
		},
		{
			name:           "invalid type input with string",
			input:          "a",
			expectedResult: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := internal.ToMapString(tt.input)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
