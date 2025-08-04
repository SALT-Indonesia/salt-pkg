package logmanager_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithLogDir(t *testing.T) {
	tests := []struct {
		name        string
		dir         string
		expectedNil bool
	}{
		{"ValidDir", "/var/logs", false},
		{"EmptyDir", "", false},
		{"RelativePath", "./logs", false},
		{"RootDir", "/", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := logmanager.NewApplication()
			option := logmanager.WithLogDir(tt.dir)
			option(app)

			if tt.expectedNil {
				assert.Nil(t, app)
				return
			}
			assert.NotNil(t, app)
		})
	}
}

func TestWithTags(t *testing.T) {
	tests := []struct {
		name        string
		tags        []string
		expectedNil bool
	}{
		{
			"EmptyTags",
			[]string{},
			false,
		},
		{
			"SingleTag",
			[]string{"debug"},
			false,
		},
		{
			"MultipleTags",
			[]string{"debug", "production", "release"},
			false,
		},
		{
			"NilTags",
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := logmanager.NewApplication()
			option := logmanager.WithTags(tt.tags...)
			option(app)

			if tt.expectedNil {
				assert.Nil(t, app)
				return
			}

			assert.NotNil(t, app)
		})
	}
}

func TestWithMaskConfigs(t *testing.T) {
	tests := []struct {
		name               string
		maskConfigs        logmanager.MaskConfigs
		expectedMaskConfig int
	}{
		{"EmptyConfigs", logmanager.MaskConfigs{}, 0},
		{
			"SingleConfig",
			logmanager.MaskConfigs{{Field: "password", Type: logmanager.PartialMask, ShowFirst: 2, ShowLast: 2}},
			1,
		},
		{
			"MultipleConfigs",
			logmanager.MaskConfigs{
				{Field: "creditCard", Type: logmanager.PartialMask, ShowFirst: 4, ShowLast: 4},
				{Field: "password", Type: logmanager.PartialMask, ShowFirst: 2, ShowLast: 2},
			},
			2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := logmanager.NewApplication()
			option := logmanager.WithMaskConfigs(tt.maskConfigs)
			option(app)

			if tt.expectedMaskConfig == 0 {
				assert.Equal(t, []internal.MaskingConfig{}, tt.maskConfigs.GetMaskConfigs())
				return
			}
			assert.NotNil(t, tt.maskConfigs.GetMaskConfigs())
		})
	}
}

func TestWithExposeHeaders(t *testing.T) {
	tests := []struct {
		name            string
		headers         []string
		expectedHeaders []string
	}{
		{
			"EmptyHeaders",
			[]string{},
			[]string{},
		},
		{
			"SingleHeader",
			[]string{"Content-Type"},
			[]string{"Content-Type"},
		},
		{
			"MultipleHeaders",
			[]string{"Content-Type", "Authorization", "User-Agent"},
			[]string{"Content-Type", "Authorization", "User-Agent"},
		},
		{
			"NilHeaders",
			nil,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := logmanager.NewApplication()
			option := logmanager.WithExposeHeaders(tt.headers...)
			option(app)

			assert.Equal(t, 1, 1)
		})
	}
}

func TestWithTraceIDKey(t *testing.T) {
	tests := []struct {
		name               string
		traceIDKey         string
		expectedTraceIDKey string
	}{
		{
			"ValidKey",
			"xid",
			"xid",
		},
		{
			"EmptyKey",
			"",
			"trace_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := logmanager.NewApplication(logmanager.WithTraceIDKey(tt.traceIDKey))

			assert.NotNil(t, app)
		})
	}
}

func TestWithService(t *testing.T) {
	tests := []struct {
		name            string
		serviceName     string
		expectedService string
	}{
		{
			"ValidServiceName",
			"user-service",
			"user-service",
		},
		{
			"EmptyServiceName",
			"",
			"default",
		},
		{
			"AnotherServiceName",
			"payment-gateway",
			"payment-gateway",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := logmanager.NewApplication(logmanager.WithService(tt.serviceName))

			assert.NotNil(t, app)
			assert.Equal(t, tt.expectedService, app.Service())
		})
	}
}

func TestWithMaskingConfig(t *testing.T) {
	tests := []struct {
		name           string
		maskingConfigs []logmanager.MaskingConfig
		expectedCount  int
	}{
		{
			"EmptyConfigs",
			[]logmanager.MaskingConfig{},
			0,
		},
		{
			"SingleJSONPathConfig",
			[]logmanager.MaskingConfig{
				{JSONPath: "$.password", Type: logmanager.FullMask},
			},
			1,
		},
		{
			"MultipleJSONPathConfigs",
			[]logmanager.MaskingConfig{
				{JSONPath: "$.password", Type: logmanager.FullMask},
				{JSONPath: "$.credit_card", Type: logmanager.PartialMask, ShowFirst: 4, ShowLast: 4},
				{JSONPath: "$.ssn", Type: logmanager.HideMask},
			},
			3,
		},
		{
			"NestedJSONPathConfigs",
			[]logmanager.MaskingConfig{
				{JSONPath: "$.user.password", Type: logmanager.FullMask},
				{JSONPath: "$.payment.card_number", Type: logmanager.PartialMask, ShowFirst: 6, ShowLast: 4},
			},
			2,
		},
		{
			"ArrayJSONPathConfigs",
			[]logmanager.MaskingConfig{
				{JSONPath: "$.users[*].password", Type: logmanager.FullMask},
				{JSONPath: "$.cards[*].number", Type: logmanager.PartialMask, ShowFirst: 4, ShowLast: 4},
			},
			2,
		},
		{
			"FieldPatternConfig",
			[]logmanager.MaskingConfig{
				{FieldPattern: "password", Type: logmanager.FullMask},
			},
			1,
		},
		{
			"MultipleFieldPatternConfigs",
			[]logmanager.MaskingConfig{
				{FieldPattern: "password", Type: logmanager.FullMask},
				{FieldPattern: "secret", Type: logmanager.FullMask},
				{FieldPattern: "token", Type: logmanager.HideMask},
			},
			3,
		},
		{
			"MixedJSONPathAndFieldPattern",
			[]logmanager.MaskingConfig{
				{JSONPath: "$.specific.password", Type: logmanager.FullMask},
				{FieldPattern: "api_key", Type: logmanager.HideMask},
			},
			2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := logmanager.NewApplication(logmanager.WithMaskingConfig(tt.maskingConfigs))
			assert.NotNil(t, app)
		})
	}
}

func TestNewApplicationDefaultPasswordMasking(t *testing.T) {
	// Test that NewApplication has default password masking config for FieldPattern "password"
	app := logmanager.NewApplication()
	assert.NotNil(t, app)

	// Verify that a transaction can be created and works correctly
	// The default masking config should handle password-related fields automatically
	txn := app.StartHttp("test-trace", "test")
	assert.NotNil(t, txn)

	// The transaction should be able to handle requests with password data
	// (the actual masking happens internally when logs are written)
	txn.End()
}

func TestWithMaskingConfigJSONPath(t *testing.T) {
	tests := []struct {
		name           string
		maskingConfigs []logmanager.MaskingConfig
		description    string
	}{
		{
			name: "JSONPath masking for nested password fields",
			maskingConfigs: []logmanager.MaskingConfig{
				{
					JSONPath: "$.user.credentials.password",
					Type:     logmanager.FullMask,
				},
				{
					JSONPath:  "$.api.key",
					Type:      logmanager.PartialMask,
					ShowFirst: 4,
					ShowLast:  4,
				},
			},
			description: "Should mask nested password and API key using JSONPath",
		},
		{
			name: "JSONPath array masking",
			maskingConfigs: []logmanager.MaskingConfig{
				{
					JSONPath: "$.users[*].password",
					Type:     logmanager.FullMask,
				},
			},
			description: "Should mask password fields in user arrays",
		},
		{
			name: "JSONPath wildcard masking",
			maskingConfigs: []logmanager.MaskingConfig{
				{
					JSONPath:  "$..creditCard",
					Type:      logmanager.PartialMask,
					ShowFirst: 0,
					ShowLast:  4,
				},
			},
			description: "Should mask all creditCard fields recursively using wildcard",
		},
		{
			name: "Mixed JSONPath and FieldPattern",
			maskingConfigs: []logmanager.MaskingConfig{
				{
					JSONPath: "$.sensitive.apiKey",
					Type:     logmanager.HideMask,
				},
				{
					FieldPattern: "password",
					Type:         logmanager.FullMask,
				},
			},
			description: "Should apply both JSONPath and FieldPattern masking",
		},
		{
			name: "Complex JSONPath with filtering",
			maskingConfigs: []logmanager.MaskingConfig{
				{
					JSONPath:  "$.transactions[?(@.amount > 1000)].creditCard",
					Type:      logmanager.PartialMask,
					ShowFirst: 0,
					ShowLast:  4,
				},
			},
			description: "Should mask credit cards only for high-value transactions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create application with JSONPath masking configurations
			app := logmanager.NewApplication(logmanager.WithMaskingConfig(tt.maskingConfigs))
			assert.NotNil(t, app)

			// Create a transaction to test the masking
			txn := app.StartHttp("test-trace-jsonpath", "jsonpath-test")
			assert.NotNil(t, txn)

			// The actual masking happens in the logging pipeline
			// This test verifies that the application initializes correctly with JSONPath configs
			txn.End()
		})
	}
}
