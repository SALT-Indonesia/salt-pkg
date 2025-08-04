package clientmanager

import (
	"net"
	"net/http"
	"time"
)

var (
	newTransport = func() *http.Transport {
		return &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          1000,                   // Global idle connections
			MaxIdleConnsPerHost:   500,                    // Idle per host, increase for burst reuse
			MaxConnsPerHost:       1000,                   // Concurrent connections per host
			IdleConnTimeout:       90 * time.Second,       // Time an idle connection remains open
			ForceAttemptHTTP2:     true,                   // Better multiplexing
			TLSHandshakeTimeout:   3 * time.Second,        // Lower TLS delay
			ExpectContinueTimeout: 500 * time.Millisecond, // Reduce delay on 100-continue
			ResponseHeaderTimeout: 5 * time.Second,        // Avoid hanging requests
			DialContext: (&net.Dialer{
				Timeout:   2 * time.Second,  // Faster failure for unreachable services
				KeepAlive: 60 * time.Second, // Better for long-lived idle pools
				DualStack: true,             // IPv4 + IPv6 if available
			}).DialContext,
		}
	}
	newClient = func() *http.Client {
		return &http.Client{
			Transport: newTransport(),
			Timeout:   10 * time.Second, // Reduce per-request timeout to catch slow services
		}
	}

	// The client has a default configuration to handle 100 concurrent requests with 10 milliseconds for each request.
	// Refer to `client_test.go`.
	client = newClient()
)
