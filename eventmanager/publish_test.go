package eventmanager

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type message struct {
	Title       string `validate:"required"`
	Description string `validate:"required"`
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name: "valid message",
			input: &message{
				Title:       "Valid Title",
				Description: "Valid Description",
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: false,
		},
		{
			name: "missing description",
			input: &message{
				Title: "Valid Title",
			},
			wantErr: true,
		},
		{
			name: "missing title",
			input: &message{
				Description: "Valid Description",
			},
			wantErr: true,
		},
		{
			name:    "unsupported type",
			input:   "this is a string",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setVars() // Initialize validator before testing
			err := validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
