package internal_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestHasErrorInternalFromHttpStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   bool
	}{
		{
			name:       "acceptable status code",
			statusCode: http.StatusOK,
			expected:   false,
		},
		{
			name:       "warning status code",
			statusCode: http.StatusAlreadyReported,
			expected:   false,
		},
		{
			name:       "unacceptable status code",
			statusCode: http.StatusInternalServerError,
			expected:   true,
		},
		{
			name:       "non-standard status code 600",
			statusCode: 600,
			expected:   true,
		},
		{
			name:       "non-standard acceptable code 299",
			statusCode: 299,
			expected:   false,
		},
		{
			name:       "temporary redirect status code",
			statusCode: http.StatusTemporaryRedirect,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := internal.HasErrorInternalFromHttpStatusCode(tt.statusCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasErrorBusinessFromHttpStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   bool
	}{
		{
			name:       "bad request status code",
			statusCode: http.StatusBadRequest,
			expected:   true,
		},
		{
			name:       "acceptable status code",
			statusCode: http.StatusOK,
			expected:   false,
		},
		{
			name:       "warning status code",
			statusCode: http.StatusAlreadyReported,
			expected:   false,
		},
		{
			name:       "unacceptable and non-warning status code",
			statusCode: http.StatusInternalServerError,
			expected:   false,
		},
		{
			name:       "non-standard status code 600",
			statusCode: 600,
			expected:   false,
		},
		{
			name:       "non-standard warning status code 299",
			statusCode: 299,
			expected:   false,
		},
		{
			name:       "temporary redirect status code",
			statusCode: http.StatusTemporaryRedirect,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := internal.HasErrorBusinessFromHttpStatusCode(tt.statusCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}
