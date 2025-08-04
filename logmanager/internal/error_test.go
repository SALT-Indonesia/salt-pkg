package internal_test

import (
	"errors"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorToString(t *testing.T) {
	tests := []struct {
		name  string
		input error
		want  string
	}{
		{
			name:  "nil error",
			input: nil,
			want:  "",
		},
		{
			name:  "valid error message",
			input: errors.New("sample error"),
			want:  "sample error",
		},
		{
			name:  "another valid error",
			input: errors.New("another error message"),
			want:  "another error message",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := internal.ErrorToString(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}
