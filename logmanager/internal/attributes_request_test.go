package internal_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestBodyConsumerAttributes(t *testing.T) {
	tests := []struct {
		name       string
		attributes *internal.Attributes
		bodyBytes  []byte
		wantNil    bool
	}{
		{
			name:       "NilBodyBytes",
			attributes: internal.NewAttributes(),
			bodyBytes:  nil,
			wantNil:    true,
		},
		{
			name:       "EmptyBodyBytes",
			attributes: internal.NewAttributes(),
			bodyBytes:  []byte{},
			wantNil:    true,
		},
		{
			name:       "ValidBodyBytes",
			attributes: internal.NewAttributes(),
			bodyBytes:  []byte(`{"key": "value"}`),
			wantNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			internal.RequestBodyConsumerAttributes(tt.attributes, tt.bodyBytes)

			if tt.wantNil {
				assert.Nil(t, tt.attributes.Value().Get(internal.AttributeConsumerRequestBody))
				return
			}
			assert.NotNil(t, tt.attributes.Value().Get(internal.AttributeConsumerRequestBody))
		})
	}
}
