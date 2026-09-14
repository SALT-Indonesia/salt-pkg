package otel

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestResolveProtocol_ExplicitConfiguredValueWins tests that an explicitly
// configured protocol takes precedence over environment variables and the
// endpoint scheme.
func TestResolveProtocol_ExplicitConfiguredValueWins(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")

	got := resolveProtocol(ProtocolHTTPProtobuf, "https://collector:4318")

	if got != ProtocolHTTPProtobuf {
		t.Errorf("expected protocol %q, got %q", ProtocolHTTPProtobuf, got)
	}
}

// TestResolveProtocol_EnvVarOTELExporterOTLPProtocol tests that the generic
// OTEL_EXPORTER_OTLP_PROTOCOL environment variable is honored when no
// explicit protocol is configured.
func TestResolveProtocol_EnvVarOTELExporterOTLPProtocol(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")

	got := resolveProtocol("", "localhost:4317")

	if got != ProtocolHTTPProtobuf {
		t.Errorf("expected protocol %q, got %q", ProtocolHTTPProtobuf, got)
	}
}

// TestResolveProtocol_TracesProtocolEnvTakesPrecedenceOverGeneric tests that
// OTEL_EXPORTER_OTLP_TRACES_PROTOCOL takes precedence over
// OTEL_EXPORTER_OTLP_PROTOCOL.
func TestResolveProtocol_TracesProtocolEnvTakesPrecedenceOverGeneric(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "grpc")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")

	got := resolveProtocol("", "localhost:4317")

	if got != ProtocolGRPC {
		t.Errorf("expected protocol %q, got %q", ProtocolGRPC, got)
	}
}

// TestResolveProtocol_EndpointSchemeInferenceHTTPS tests that an https://
// endpoint scheme is inferred as http/protobuf when nothing else is set.
func TestResolveProtocol_EndpointSchemeInferenceHTTPS(t *testing.T) {
	got := resolveProtocol("", "https://collector:4318")

	if got != ProtocolHTTPProtobuf {
		t.Errorf("expected protocol %q, got %q", ProtocolHTTPProtobuf, got)
	}
}

// TestResolveProtocol_EndpointSchemeInferenceHTTP tests that an http://
// endpoint scheme is inferred as http/protobuf when nothing else is set.
func TestResolveProtocol_EndpointSchemeInferenceHTTP(t *testing.T) {
	got := resolveProtocol("", "http://collector:4318")

	if got != ProtocolHTTPProtobuf {
		t.Errorf("expected protocol %q, got %q", ProtocolHTTPProtobuf, got)
	}
}

// TestResolveProtocol_DefaultsToGRPC tests that grpc is the default when no
// protocol, env var, or recognizable endpoint scheme is present.
func TestResolveProtocol_DefaultsToGRPC(t *testing.T) {
	got := resolveProtocol("", "localhost:4317")

	if got != ProtocolGRPC {
		t.Errorf("expected protocol %q, got %q", ProtocolGRPC, got)
	}
}

// TestNormalizeProtocol_UnknownValueReturnsEmpty tests that an unrecognized
// or empty protocol string normalizes to an empty string.
func TestNormalizeProtocol_UnknownValueReturnsEmpty(t *testing.T) {
	if got := normalizeProtocol("carrier-pigeon"); got != "" {
		t.Errorf("expected empty string for unknown protocol, got %q", got)
	}
	if got := normalizeProtocol(""); got != "" {
		t.Errorf("expected empty string for empty protocol, got %q", got)
	}
}

// generateSelfSignedCAPEM creates a minimal self-signed CA certificate in
// PEM format for use in TLS-config-loading tests.
func generateSelfSignedCAPEM(t *testing.T) []byte {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

// TestLoadCATLSConfig_ValidPEM tests that a valid PEM CA certificate file
// produces a usable *tls.Config with the CA in its root pool.
func TestLoadCATLSConfig_ValidPEM(t *testing.T) {
	certPEM := generateSelfSignedCAPEM(t)
	certPath := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}

	tlsConfig, err := loadCATLSConfig(certPath)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tlsConfig == nil {
		t.Fatal("expected non-nil tls.Config")
	}
	if tlsConfig.RootCAs == nil {
		t.Fatal("expected RootCAs to be set")
	}
}

// TestLoadCATLSConfig_MissingFileReturnsError tests that a nonexistent
// certificate path returns an error.
func TestLoadCATLSConfig_MissingFileReturnsError(t *testing.T) {
	_, err := loadCATLSConfig(filepath.Join(t.TempDir(), "does-not-exist.pem"))

	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// TestLoadCATLSConfig_InvalidPEMReturnsError tests that a file with no valid
// PEM certificates returns an error.
func TestLoadCATLSConfig_InvalidPEMReturnsError(t *testing.T) {
	certPath := filepath.Join(t.TempDir(), "invalid.pem")
	if err := os.WriteFile(certPath, []byte("not a certificate"), 0o600); err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}

	_, err := loadCATLSConfig(certPath)

	if err == nil {
		t.Fatal("expected error for invalid PEM content, got nil")
	}
}
