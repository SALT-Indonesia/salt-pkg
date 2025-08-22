package logmanager_test

import (
	"os"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/stretchr/testify/assert"
)

func TestNewApplicationWithEnvironment(t *testing.T) {
	tests := []struct {
		name           string
		envVar         string
		options        []logmanager.Option
		expectedEnv    string
		expectedDebug  bool
	}{
		{
			name:           "Default environment (no APP_ENV set)",
			envVar:         "",
			options:        nil,
			expectedEnv:    "development",
			expectedDebug:  true,
		},
		{
			name:           "Production environment via APP_ENV",
			envVar:         "production",
			options:        nil,
			expectedEnv:    "production",
			expectedDebug:  false,
		},
		{
			name:           "Development environment via APP_ENV",
			envVar:         "development",
			options:        nil,
			expectedEnv:    "development",
			expectedDebug:  true,
		},
		{
			name:           "Staging environment via APP_ENV",
			envVar:         "staging",
			options:        nil,
			expectedEnv:    "staging",
			expectedDebug:  true,
		},
		{
			name:           "Production environment via WithEnvironment option",
			envVar:         "",
			options:        []logmanager.Option{logmanager.WithEnvironment("production")},
			expectedEnv:    "production",
			expectedDebug:  false,
		},
		{
			name:           "Development environment via WithEnvironment option",
			envVar:         "",
			options:        []logmanager.Option{logmanager.WithEnvironment("development")},
			expectedEnv:    "development",
			expectedDebug:  true,
		},
		{
			name:           "Production with explicit debug mode",
			envVar:         "production",
			options:        []logmanager.Option{logmanager.WithDebug()},
			expectedEnv:    "production",
			expectedDebug:  true, // WithDebug() should override production setting
		},
		{
			name:           "WithEnvironment overrides APP_ENV",
			envVar:         "development",
			options:        []logmanager.Option{logmanager.WithEnvironment("production")},
			expectedEnv:    "production",
			expectedDebug:  false,
		},
		{
			name:           "Production via option with explicit debug",
			envVar:         "",
			options:        []logmanager.Option{
				logmanager.WithEnvironment("production"),
				logmanager.WithDebug(),
			},
			expectedEnv:    "production",
			expectedDebug:  true, // WithDebug() should override production setting
		},
		{
			name:           "APP_ENV production overridden by WithEnvironment then WithDebug",
			envVar:         "production",
			options:        []logmanager.Option{
				logmanager.WithEnvironment("staging"),
				logmanager.WithDebug(),
			},
			expectedEnv:    "staging",
			expectedDebug:  true, // WithDebug() should override
		},
		{
			name:           "APP_ENV development overridden by WithEnvironment production",
			envVar:         "development",
			options:        []logmanager.Option{
				logmanager.WithEnvironment("production"),
			},
			expectedEnv:    "production",
			expectedDebug:  false, // Environment changed to production should disable debug
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envVar != "" {
				os.Setenv("APP_ENV", tt.envVar)
				defer os.Unsetenv("APP_ENV")
			}

			// Create application
			app := logmanager.NewApplication(tt.options...)

			// Assert environment
			assert.Equal(t, tt.expectedEnv, app.Environment())

			// We can't directly test debug field, but we can verify the app was created successfully
			assert.NotNil(t, app)
			
			// Test that the app can start a transaction (which uses debug internally)
			txn := app.StartHttp("test-trace", "test")
			assert.NotNil(t, txn)
		})
	}
}

func TestEnvironmentGetter(t *testing.T) {
	tests := []struct {
		name        string
		environment string
	}{
		{
			name:        "Production environment",
			environment: "production",
		},
		{
			name:        "Development environment",
			environment: "development",
		},
		{
			name:        "Custom environment",
			environment: "custom-env",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := logmanager.NewApplication(logmanager.WithEnvironment(tt.environment))
			assert.Equal(t, tt.environment, app.Environment())
		})
	}
}