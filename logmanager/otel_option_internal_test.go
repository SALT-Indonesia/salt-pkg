package logmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildOTelConfig_Defaults(t *testing.T) {
	cfg := buildOTelConfig("my-service", "production", nil)

	assert.Equal(t, "localhost:4317", cfg.Endpoint)
	assert.True(t, cfg.Insecure)
	assert.Equal(t, "my-service", cfg.ServiceName)
	assert.Equal(t, "production", cfg.Environment)
	assert.Empty(t, cfg.Protocol)
	assert.Empty(t, cfg.CertificatePath)
	assert.False(t, cfg.FromEnv)
}

func TestBuildOTelConfig_WithOTelProtocol(t *testing.T) {
	cfg := buildOTelConfig("my-service", "production", []OTelExporterOption{
		WithOTelProtocol("http/protobuf"),
	})

	assert.Equal(t, "http/protobuf", cfg.Protocol)
}

func TestBuildOTelConfig_WithOTelCertificate_SetsCertPathAndClearsInsecure(t *testing.T) {
	cfg := buildOTelConfig("my-service", "production", []OTelExporterOption{
		WithOTelInsecure(),
		WithOTelCertificate("/etc/ssl/certs/ca.pem"),
	})

	assert.Equal(t, "/etc/ssl/certs/ca.pem", cfg.CertificatePath)
	assert.False(t, cfg.Insecure)
}

func TestBuildOTelConfig_WithOTelFromEnv_SetsFromEnvTrue(t *testing.T) {
	cfg := buildOTelConfig("my-service", "production", []OTelExporterOption{
		WithOTelFromEnv(),
	})

	assert.True(t, cfg.FromEnv)
}
