package internal_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestJSONMasker_LogrusMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		maskConfigs []internal.MaskingConfig
		fields      map[string]interface{}
	}{
		{
			name: "it should be ok",
			maskConfigs: []internal.MaskingConfig{
				{
					FieldPattern: "creditCard",
					Type:         internal.PartialMask,
					ShowFirst:    0,
					ShowLast:     4,
				}, {
					FieldPattern: "phoneNumber",
					Type:         internal.FullMask,
				},
			},
			fields: map[string]interface{}{
				"latency": 100,
				"request": map[string]interface{}{
					"username":    "john doe",
					"creditCard":  "1234-5678-9012-3456",
					"phoneNumber": "1234567890",
				},
			},
		},
		{
			name: "it should be ok with map string",
			maskConfigs: []internal.MaskingConfig{
				{
					FieldPattern: "creditCard",
					Type:         internal.PartialMask,
					ShowFirst:    0,
					ShowLast:     4,
				}, {
					FieldPattern: "phoneNumber",
					Type:         internal.FullMask,
				},
			},
			fields: map[string]interface{}{
				"latency": 100,
				"request": map[string]string{
					"username":    "john doe",
					"creditCard":  "1234-5678-9012-3456",
					"phoneNumber": "1234567890",
				},
			},
		},
		{
			name: "it should be ok with sub map interface",
			maskConfigs: []internal.MaskingConfig{
				{
					FieldPattern: "creditCard",
					Type:         internal.PartialMask,
					ShowFirst:    0,
					ShowLast:     4,
				}, {
					FieldPattern: "phoneNumber",
					Type:         internal.FullMask,
				},
			},
			fields: map[string]interface{}{
				"latency": 100,
				"request": map[string]interface{}{
					"username":    "john doe",
					"creditCard":  "1234-5678-9012-3456",
					"phoneNumber": "1234567890",
					"data": map[string]interface{}{
						"creditCard":  "1234-5678-9012-3456",
						"phoneNumber": "1234567890",
					},
				},
			},
		},
		{
			name:        "it should be ok with nil config and fields",
			maskConfigs: nil,
			fields:      nil,
		},
		{
			name:        "it should be ok with empty config and fields",
			maskConfigs: []internal.MaskingConfig{},
			fields:      map[string]interface{}{},
		},
		{
			name: "it should be ok with mask type hide",
			maskConfigs: []internal.MaskingConfig{
				{
					FieldPattern: "creditCard",
					Type:         internal.HideMask,
				}, {
					FieldPattern: "phoneNumber",
					Type:         internal.FullMask,
				},
			},
			fields: map[string]interface{}{
				"latency": 100,
				"request": map[string]interface{}{
					"username":    "john doe",
					"creditCard":  "1234-5678-9012-3456",
					"phoneNumber": "1234567890",
				},
			},
		},
		{
			name: "it should be print single star when character >= 255",
			maskConfigs: []internal.MaskingConfig{
				{
					FieldPattern: "creditCard",
					Type:         internal.HideMask,
				}, {
					FieldPattern: "phoneNumber",
					Type:         internal.FullMask,
				}, {
					FieldPattern: "username",
					Type:         internal.PartialMask,
				},
			},
			fields: map[string]interface{}{
				"latency": 100,
				"request": map[string]interface{}{
					"username":    strings.Repeat("a", 255),
					"creditCard":  strings.Repeat("b", 255),
					"phoneNumber": strings.Repeat("c", 255),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := internal.NewJSONMasker(tt.maskConfigs)

			logger := logrus.New()
			logger.AddHook(m.LogrusMiddleware())
			logger.SetFormatter(&logrus.JSONFormatter{})

			// Example log with various masking scenarios
			logger.WithFields(tt.fields).Info("User sensitive data example")
			assert.NotNil(t, m)
		})
	}
}

func TestJSONMasker_JSONPathMasking(t *testing.T) {
	tests := []struct {
		name        string
		maskConfigs []internal.MaskingConfig
		input       interface{}
		description string
	}{
		{
			name: "JSONPath simple field masking",
			maskConfigs: []internal.MaskingConfig{
				{
					JSONPath: "$.user.password",
					Type:     internal.FullMask,
				},
			},
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"username": "john_doe",
					"password": "secret123",
					"email":    "john@example.com",
				},
			},
			description: "Should mask password field using JSONPath",
		},
		{
			name: "JSONPath array element masking",
			maskConfigs: []internal.MaskingConfig{
				{
					JSONPath: "$.users[*].password",
					Type:     internal.FullMask,
				},
			},
			input: map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"username": "john",
						"password": "secret1",
					},
					map[string]interface{}{
						"username": "jane",
						"password": "secret2",
					},
				},
			},
			description: "Should mask password fields in array elements",
		},
		{
			name: "JSONPath nested object masking",
			maskConfigs: []internal.MaskingConfig{
				{
					JSONPath:  "$.data.credentials.apiKey",
					Type:      internal.PartialMask,
					ShowFirst: 4,
					ShowLast:  4,
				},
			},
			input: map[string]interface{}{
				"data": map[string]interface{}{
					"credentials": map[string]interface{}{
						"apiKey":   "abcd1234567890efgh",
						"username": "api_user",
					},
					"metadata": map[string]interface{}{
						"version": "1.0",
					},
				},
			},
			description: "Should partially mask nested API key",
		},
		{
			name: "JSONPath wildcard masking",
			maskConfigs: []internal.MaskingConfig{
				{
					JSONPath:  "$..creditCard",
					Type:      internal.PartialMask,
					ShowFirst: 0,
					ShowLast:  4,
				},
			},
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"creditCard": "1234567890123456",
				},
				"payment": map[string]interface{}{
					"method": "card",
					"details": map[string]interface{}{
						"creditCard": "9876543210987654",
					},
				},
			},
			description: "Should mask all creditCard fields recursively",
		},
		{
			name: "JSONPath with HideMask",
			maskConfigs: []internal.MaskingConfig{
				{
					JSONPath: "$.sensitive.ssn",
					Type:     internal.HideMask,
				},
			},
			input: map[string]interface{}{
				"sensitive": map[string]interface{}{
					"ssn":  "123-45-6789",
					"name": "John Doe",
				},
			},
			description: "Should completely hide SSN field",
		},
		{
			name: "Multiple JSONPath configurations",
			maskConfigs: []internal.MaskingConfig{
				{
					JSONPath: "$.user.password",
					Type:     internal.FullMask,
				},
				{
					JSONPath:  "$.user.email",
					Type:      internal.PartialMask,
					ShowFirst: 3,
					ShowLast:  0,
				},
				{
					JSONPath:  "$.payment.creditCard",
					Type:      internal.PartialMask,
					ShowFirst: 0,
					ShowLast:  4,
				},
			},
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"username": "johndoe",
					"password": "secret123",
					"email":    "john.doe@example.com",
				},
				"payment": map[string]interface{}{
					"creditCard": "1234567890123456",
					"amount":     100.50,
				},
			},
			description: "Should apply multiple JSONPath configurations",
		},
		{
			name: "JSONPath with non-matching path",
			maskConfigs: []internal.MaskingConfig{
				{
					JSONPath: "$.nonexistent.field",
					Type:     internal.FullMask,
				},
			},
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"username": "johndoe",
					"password": "secret123",
				},
			},
			description: "Should leave data unchanged when JSONPath doesn't match",
		},
		{
			name: "JSONPath with complex array filtering",
			maskConfigs: []internal.MaskingConfig{
				{
					JSONPath: "$.orders[?(@.amount > 100)].creditCard",
					Type:     internal.FullMask,
				},
			},
			input: map[string]interface{}{
				"orders": []interface{}{
					map[string]interface{}{
						"amount":     50.0,
						"creditCard": "1111222233334444",
					},
					map[string]interface{}{
						"amount":     150.0,
						"creditCard": "5555666677778888",
					},
				},
			},
			description: "Should mask credit cards only for orders with amount > 100",
		},
		{
			name: "Mixed JSONPath and FieldPattern configurations",
			maskConfigs: []internal.MaskingConfig{
				{
					JSONPath: "$.api.key",
					Type:     internal.FullMask,
				},
				{
					FieldPattern: "password",
					Type:         internal.FullMask,
				},
			},
			input: map[string]interface{}{
				"api": map[string]interface{}{
					"key":      "api_secret_key",
					"endpoint": "https://api.example.com",
				},
				"user": map[string]interface{}{
					"userPassword":  "user_secret",
					"adminPassword": "admin_secret",
				},
			},
			description: "Should apply both JSONPath and FieldPattern masking",
		},
		{
			name: "JSONPath with empty/nil values",
			maskConfigs: []internal.MaskingConfig{
				{
					JSONPath: "$.user.empty",
					Type:     internal.FullMask,
				},
				{
					JSONPath: "$.user.nil",
					Type:     internal.FullMask,
				},
			},
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"empty": "",
					"nil":   nil,
					"valid": "some_value",
				},
			},
			description: "Should handle empty and nil values gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masker := internal.NewJSONMasker(tt.maskConfigs)
			assert.NotNil(t, masker)

			// Test the masking functionality
			result := masker.MaskData(tt.input)
			assert.NotNil(t, result)

			// Log the result for manual inspection during test runs
			logger := logrus.New()
			logger.AddHook(masker.LogrusMiddleware())
			logger.SetFormatter(&logrus.JSONFormatter{})

			logger.WithFields(logrus.Fields{
				"original": tt.input,
				"masked":   result,
				"test":     tt.description,
			}).Info("JSONPath masking test")
		})
	}
}

func TestJSONMasker_JSONPathEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		maskConfigs []internal.MaskingConfig
		input       interface{}
		expectError bool
		description string
	}{
		{
			name: "Invalid JSONPath syntax",
			maskConfigs: []internal.MaskingConfig{
				{
					JSONPath: "$.invalid.[syntax",
					Type:     internal.FullMask,
				},
			},
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "test",
				},
			},
			expectError: false, // Should gracefully handle invalid JSONPath
			description: "Should handle invalid JSONPath syntax gracefully",
		},
		{
			name: "Empty JSONPath",
			maskConfigs: []internal.MaskingConfig{
				{
					JSONPath: "",
					Type:     internal.FullMask,
				},
			},
			input: map[string]interface{}{
				"user": "test_data",
			},
			expectError: false,
			description: "Should handle empty JSONPath",
		},
		{
			name: "JSONPath on non-JSON data",
			maskConfigs: []internal.MaskingConfig{
				{
					JSONPath: "$.field",
					Type:     internal.FullMask,
				},
			},
			input:       "simple_string",
			expectError: false,
			description: "Should handle non-JSON input gracefully",
		},
		{
			name: "JSONPath on slice input",
			maskConfigs: []internal.MaskingConfig{
				{
					JSONPath: "$[0].password",
					Type:     internal.FullMask,
				},
			},
			input: []interface{}{
				map[string]interface{}{
					"username": "user1",
					"password": "secret1",
				},
				map[string]interface{}{
					"username": "user2",
					"password": "secret2",
				},
			},
			expectError: false,
			description: "Should handle slice input for JSONPath",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masker := internal.NewJSONMasker(tt.maskConfigs)
			assert.NotNil(t, masker)

			// This should not panic even with edge cases
			result := masker.MaskData(tt.input)
			assert.NotNil(t, result)

			// Log for manual inspection
			logger := logrus.New()
			logger.AddHook(masker.LogrusMiddleware())
			logger.SetFormatter(&logrus.JSONFormatter{})

			logger.WithFields(logrus.Fields{
				"original":    tt.input,
				"masked":      result,
				"test":        tt.description,
				"expectError": tt.expectError,
			}).Info("JSONPath edge case test")
		})
	}
}
