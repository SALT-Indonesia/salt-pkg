package internal_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResponseBodyAttribute(t *testing.T) {
	tests := []struct {
		name       string
		attributes *internal.Attributes
		bodyBytes  []byte
		expected   map[string]interface{}
	}{
		{
			name:       "EmptyBody",
			attributes: internal.NewAttributes(),
			bodyBytes:  []byte(""),
			expected:   map[string]interface{}{},
		},
		{
			name:       "SimpleString",
			attributes: internal.NewAttributes(),
			bodyBytes:  []byte(`{"simple": "string"}`),
			expected: map[string]interface{}{
				internal.AttributeResponseBody: map[string]interface{}{
					"simple": "string",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			internal.ResponseBodyAttribute(tt.attributes, tt.bodyBytes)
			assert.Equal(t, tt.expected, tt.attributes.Value().Values())
		})
	}
}
