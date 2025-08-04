package logmanager_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMaskConfigs(t *testing.T) {
	tests := []struct {
		name     string
		input    logmanager.MaskConfigs
		expected []internal.MaskingConfig
	}{
		{
			name:     "empty input",
			input:    logmanager.MaskConfigs{},
			expected: []internal.MaskingConfig{},
		},
		{
			name: "single config",
			input: logmanager.MaskConfigs{
				{
					Field: "field1",
					Type:  logmanager.FullMask,
				},
			},
			expected: []internal.MaskingConfig{
				{
					Field: "field1",
					Type:  logmanager.FullMask,
				},
			},
		},
		{
			name: "multiple configs",
			input: logmanager.MaskConfigs{
				{
					Field:     "field1",
					Type:      logmanager.PartialMask,
					ShowFirst: 2,
					ShowLast:  3,
				},
				{
					Field: "field2",
					Type:  logmanager.HideMask,
				},
			},
			expected: []internal.MaskingConfig{
				{
					Field:     "field1",
					Type:      logmanager.PartialMask,
					ShowFirst: 2,
					ShowLast:  3,
				},
				{
					Field: "field2",
					Type:  logmanager.HideMask,
				},
			},
		},
		{
			name: "default values",
			input: logmanager.MaskConfigs{
				{
					Field: "field3",
				},
			},
			expected: []internal.MaskingConfig{
				{
					Field: "field3",
					Type:  logmanager.FullMask,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.input.GetMaskConfigs()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
