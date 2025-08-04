package httpmanager

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultOption(t *testing.T) {
	opt := newDefaultOption()

	assert.NotNil(t, opt, "Default option should not be nil")
	assert.Equal(t, ":8080", opt.addr, "Default address should be :8080")
	assert.Equal(t, 10*time.Second, opt.readTimeout, "Default read timeout should be 10 seconds")
	assert.Equal(t, 10*time.Second, opt.writeTimeout, "Default write timeout should be 10 seconds")
	assert.False(t, opt.ssl, "Default SSL should be false")
	assert.Empty(t, opt.certFile, "Default certFile should be empty")
	assert.Empty(t, opt.keyFile, "Default keyFile should be empty")
	assert.Empty(t, opt.certData, "Default certData should be empty")
	assert.Empty(t, opt.keyData, "Default keyData should be empty")
}

func TestWithAddr(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected string
	}{
		{
			name:     "empty address",
			addr:     "",
			expected: "",
		},
		{
			name:     "custom address",
			addr:     ":9090",
			expected: ":9090",
		},
		{
			name:     "localhost address",
			addr:     "localhost:8080",
			expected: "localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{}
			optFunc := WithAddr(tt.addr)
			optFunc(opt)

			assert.Equal(t, tt.expected, opt.addr, "Address should be set correctly")
		})
	}
}

func TestWithReadTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{
			name:     "zero timeout",
			timeout:  0,
			expected: 0,
		},
		{
			name:     "5 seconds timeout",
			timeout:  5 * time.Second,
			expected: 5 * time.Second,
		},
		{
			name:     "100 milliseconds timeout",
			timeout:  100 * time.Millisecond,
			expected: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{}
			optFunc := WithReadTimeout(tt.timeout)
			optFunc(opt)

			assert.Equal(t, tt.expected, opt.readTimeout, "Read timeout should be set correctly")
		})
	}
}

func TestWithWriteTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{
			name:     "zero timeout",
			timeout:  0,
			expected: 0,
		},
		{
			name:     "5 seconds timeout",
			timeout:  5 * time.Second,
			expected: 5 * time.Second,
		},
		{
			name:     "100 milliseconds timeout",
			timeout:  100 * time.Millisecond,
			expected: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{}
			optFunc := WithWriteTimeout(tt.timeout)
			optFunc(opt)

			assert.Equal(t, tt.expected, opt.writeTimeout, "Write timeout should be set correctly")
		})
	}
}

func TestWithSSL(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "enable SSL",
			enabled:  true,
			expected: true,
		},
		{
			name:     "disable SSL",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{}
			optFunc := WithSSL(tt.enabled)
			optFunc(opt)

			assert.Equal(t, tt.expected, opt.ssl, "SSL flag should be set correctly")
		})
	}
}

func TestWithCertFile(t *testing.T) {
	tests := []struct {
		name     string
		certFile string
		expected string
	}{
		{
			name:     "empty cert file",
			certFile: "",
			expected: "",
		},
		{
			name:     "valid cert file path",
			certFile: "/path/to/cert.pem",
			expected: "/path/to/cert.pem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{}
			optFunc := WithCertFile(tt.certFile)
			optFunc(opt)

			assert.Equal(t, tt.expected, opt.certFile, "Certificate file path should be set correctly")
		})
	}
}

func TestWithKeyFile(t *testing.T) {
	tests := []struct {
		name     string
		keyFile  string
		expected string
	}{
		{
			name:     "empty key file",
			keyFile:  "",
			expected: "",
		},
		{
			name:     "valid key file path",
			keyFile:  "/path/to/key.pem",
			expected: "/path/to/key.pem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{}
			optFunc := WithKeyFile(tt.keyFile)
			optFunc(opt)

			assert.Equal(t, tt.expected, opt.keyFile, "Key file path should be set correctly")
		})
	}
}

func TestWithCertData(t *testing.T) {
	tests := []struct {
		name     string
		certData string
		expected string
	}{
		{
			name:     "empty cert data",
			certData: "",
			expected: "",
		},
		{
			name:     "valid cert data",
			certData: "-----BEGIN CERTIFICATE-----\nMIICertificateContent\n-----END CERTIFICATE-----",
			expected: "-----BEGIN CERTIFICATE-----\nMIICertificateContent\n-----END CERTIFICATE-----",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{}
			optFunc := WithCertData(tt.certData)
			optFunc(opt)

			assert.Equal(t, tt.expected, opt.certData, "Certificate data should be set correctly")
		})
	}
}

func TestWithKeyData(t *testing.T) {
	tests := []struct {
		name     string
		keyData  string
		expected string
	}{
		{
			name:     "empty key data",
			keyData:  "",
			expected: "",
		},
		{
			name:     "valid key data",
			keyData:  "-----BEGIN PRIVATE KEY-----\nMIIPrivateKeyContent\n-----END PRIVATE KEY-----",
			expected: "-----BEGIN PRIVATE KEY-----\nMIIPrivateKeyContent\n-----END PRIVATE KEY-----",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{}
			optFunc := WithKeyData(tt.keyData)
			optFunc(opt)

			assert.Equal(t, tt.expected, opt.keyData, "Key data should be set correctly")
		})
	}
}

func TestWithPort(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected string
	}{
		{
			name:     "empty port",
			port:     "",
			expected: ":",
		},
		{
			name:     "numeric port",
			port:     "8080",
			expected: ":8080",
		},
		{
			name:     "custom port",
			port:     "9090",
			expected: ":9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Option{}
			optFunc := WithPort(tt.port)
			optFunc(opt)

			assert.Equal(t, tt.expected, opt.addr, "Port should be set correctly in addr field")
		})
	}
}

func TestOptionFuncChaining(t *testing.T) {
	// Test that multiple option functions can be applied to the same Option struct
	opt := &Option{}

	// Apply multiple option functions
	WithAddr(":9090")(opt)
	WithReadTimeout(5 * time.Second)(opt)
	WithWriteTimeout(7 * time.Second)(opt)
	WithSSL(true)(opt)
	WithCertFile("/path/to/cert.pem")(opt)
	WithKeyFile("/path/to/key.pem")(opt)
	WithCertData("-----BEGIN CERTIFICATE-----\nMIICertificateContent\n-----END CERTIFICATE-----")(opt)
	WithKeyData("-----BEGIN PRIVATE KEY-----\nMIIPrivateKeyContent\n-----END PRIVATE KEY-----")(opt)

	// Verify all options were applied correctly
	assert.Equal(t, ":9090", opt.addr, "Address should be set correctly")
	assert.Equal(t, 5*time.Second, opt.readTimeout, "Read timeout should be set correctly")
	assert.Equal(t, 7*time.Second, opt.writeTimeout, "Write timeout should be set correctly")
	assert.True(t, opt.ssl, "SSL should be enabled")
	assert.Equal(t, "/path/to/cert.pem", opt.certFile, "Certificate file path should be set correctly")
	assert.Equal(t, "/path/to/key.pem", opt.keyFile, "Key file path should be set correctly")
	assert.Equal(t, "-----BEGIN CERTIFICATE-----\nMIICertificateContent\n-----END CERTIFICATE-----", opt.certData, "Certificate data should be set correctly")
	assert.Equal(t, "-----BEGIN PRIVATE KEY-----\nMIIPrivateKeyContent\n-----END PRIVATE KEY-----", opt.keyData, "Key data should be set correctly")
}
