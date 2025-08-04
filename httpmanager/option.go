package httpmanager

import (
	"github.com/gorilla/mux"
	"os"
	"time"
)

type Option struct {
	addr         string
	readTimeout  time.Duration
	writeTimeout time.Duration
	ssl          bool
	certFile     string
	keyFile      string
	certData     string
	keyData      string
	middlewares  []mux.MiddlewareFunc
}

// newDefaultOption initializes an Option struct with default server configurations and returns a pointer to it.
func newDefaultOption() *Option {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return &Option{
		addr:         ":" + port,
		readTimeout:  10 * time.Second,
		writeTimeout: 10 * time.Second,
		middlewares:  []mux.MiddlewareFunc{},
	}
}

type OptionFunc func(*Option)

// WithAddr sets the address field in the Option struct.
func WithAddr(addr string) OptionFunc {
	return func(o *Option) {
		o.addr = addr
	}
}

// WithReadTimeout sets the read timeout duration for the Option configuration.
func WithReadTimeout(readTimeout time.Duration) OptionFunc {
	return func(o *Option) {
		o.readTimeout = readTimeout
	}
}

// WithWriteTimeout sets the write timeout duration for the Option configuration.
func WithWriteTimeout(writeTimeout time.Duration) OptionFunc {
	return func(o *Option) {
		o.writeTimeout = writeTimeout
	}
}

// WithSSL enables or disables SSL
func WithSSL(enabled bool) OptionFunc {
	return func(o *Option) {
		o.ssl = enabled
	}
}

// WithCertFile sets the certificate file path for SSL
func WithCertFile(certFile string) OptionFunc {
	return func(o *Option) {
		o.certFile = certFile
	}
}

// WithKeyFile sets the key file path for SSL
func WithKeyFile(keyFile string) OptionFunc {
	return func(o *Option) {
		o.keyFile = keyFile
	}
}

// WithCertData sets the certificate data as a string
func WithCertData(certData string) OptionFunc {
	return func(o *Option) {
		o.certData = certData
	}
}

// WithKeyData sets the key data as a string
func WithKeyData(keyData string) OptionFunc {
	return func(o *Option) {
		o.keyData = keyData
	}
}

// WithPort sets the port for the server address
func WithPort(port string) OptionFunc {
	return func(o *Option) {
		o.addr = ":" + port
	}
}

// WithMiddleware adds middleware to the server
func WithMiddleware(middleware ...mux.MiddlewareFunc) OptionFunc {
	return func(o *Option) {
		o.middlewares = append(o.middlewares, middleware...)
	}
}
