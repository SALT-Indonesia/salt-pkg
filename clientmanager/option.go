package clientmanager

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/Azure/go-ntlmssp"
	"github.com/icholy/digest"
)

type Option func(*callOptions)

func withClient(client *http.Client) Option {
	return func(co *callOptions) {
		co.client = client
	}
}

func WithFormURLEncoded() Option {
	return func(co *callOptions) {
		co.isFormURLEncoded = true
	}
}

func WithInsecure() Option {
	return func(co *callOptions) {
		if _, ok := co.client.Transport.(*http.Transport); ok {
			if co.client.Transport.(*http.Transport).TLSClientConfig != nil {
				co.client.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true
			} else {
				co.client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
					InsecureSkipVerify: true,
				}
			}
		}
	}
}

func WithFiles(files map[string]string) Option {
	return func(co *callOptions) {
		co.files = files
	}
}

func WithHeaders(headers http.Header) Option {
	return func(co *callOptions) {
		co.headers = headers
	}
}

func WithHost(host string) Option {
	return func(co *callOptions) {
		co.host = host
	}
}

func WithMethod(method string) Option {
	return func(co *callOptions) {
		co.method = method
	}
}

func WithRequestBody(requestBody any) Option {
	return func(co *callOptions) {
		co.requestBody = requestBody
	}
}

func WithURLValues(urlValues url.Values) Option {
	return func(co *callOptions) {
		co.urlValues = urlValues
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(co *callOptions) {
		co.client.Timeout = timeout
	}
}

// WithConnectionLimit sets connection limit.
//
// Parameters:
//   - maxIdleConns: controls the maximum number of idle (keep-alive) connections across all hosts. Zero means no limit.
//   - maxIdleConnsPerHost: if non-zero, controls the maximum idle (keep-alive) connections to keep per-host. If zero DefaultMaxIdleConnsPerHost is used.
//   - maxConnsPerHost: optionally limits the total number of connections per host, including connections in the dialing, active, and idle states. On limit violation, dials will block. Zero means no limit.
//
// Returns:
//   - Option: the option
func WithConnectionLimit(maxIdleConns, maxIdleConnsPerHost, maxConnsPerHost int) Option {
	return func(co *callOptions) {
		if _, ok := co.client.Transport.(*http.Transport); ok {
			co.client.Transport.(*http.Transport).MaxIdleConns = maxIdleConns
			co.client.Transport.(*http.Transport).MaxIdleConnsPerHost = maxIdleConnsPerHost
			co.client.Transport.(*http.Transport).MaxConnsPerHost = maxConnsPerHost
		}
	}
}

func WithIdleConnTimeout(timeout time.Duration) Option {
	return func(co *callOptions) {
		if _, ok := co.client.Transport.(*http.Transport); ok {
			co.client.Transport.(*http.Transport).IdleConnTimeout = timeout
		}
	}
}

func WithTLSHandshakeTimeout(timeout time.Duration) Option {
	return func(co *callOptions) {
		if _, ok := co.client.Transport.(*http.Transport); ok {
			co.client.Transport.(*http.Transport).TLSHandshakeTimeout = timeout
		}
	}
}

func WithExpectContinueTimeout(timeout time.Duration) Option {
	return func(co *callOptions) {
		if _, ok := co.client.Transport.(*http.Transport); ok {
			co.client.Transport.(*http.Transport).ExpectContinueTimeout = timeout
		}
	}
}

func WithDialContext(timeout, keepAlive time.Duration) Option {
	return func(co *callOptions) {
		if _, ok := co.client.Transport.(*http.Transport); ok {
			co.client.Transport.(*http.Transport).DialContext = (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: keepAlive,
			}).DialContext
		}
	}
}

func WithProxy(proxyURL string) (Option, error) {
	anURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}
	return func(co *callOptions) {
		if _, ok := co.client.Transport.(*http.Transport); ok {
			co.client.Transport.(*http.Transport).Proxy = http.ProxyURL(anURL)
		}
	}, nil
}

func WithCertificates(certificates ...tls.Certificate) Option {
	return func(co *callOptions) {
		if _, ok := co.client.Transport.(*http.Transport); ok {
			if co.client.Transport.(*http.Transport).TLSClientConfig != nil {
				co.client.Transport.(*http.Transport).TLSClientConfig.Certificates = certificates
			} else {
				co.client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
					Certificates: certificates,
				}
			}
		}
	}
}

func WithAuth(auth Auth) Option {
	return func(co *callOptions) {
		co.auth = auth
	}
}

func WithAuthDigest(username, password string) Option {
	return func(co *callOptions) {
		if _, ok := co.client.Transport.(*digest.Transport); !ok {
			co.client.Transport = &digest.Transport{
				Transport: co.client.Transport,
			}
		}
		if _, ok := co.client.Transport.(*digest.Transport); ok {
			co.client.Transport.(*digest.Transport).Username = username
			co.client.Transport.(*digest.Transport).Password = password
		}
	}
}

func WithOAuth1(params OAuth1Parameters) Option {
	return func(co *callOptions) {
		co.client = params.Client()
	}
}

func WithOAuth2[Argument OAuth2Argument](params OAuth2Parameters[Argument]) (Option, error) {
	client, err := params.Client()
	if err != nil {
		return nil, err
	}
	return func(co *callOptions) {
		co.client = client
	}, nil
}

func WithAuthNTLM(auth Auth) Option {
	return func(co *callOptions) {
		if _, ok := co.client.Transport.(ntlmssp.Negotiator); !ok {
			co.client.Transport = ntlmssp.Negotiator{
				RoundTripper: co.client.Transport,
			}
		}
		co.auth = auth
	}
}
