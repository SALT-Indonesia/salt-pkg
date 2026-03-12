package clientmanager

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/Azure/go-ntlmssp"
	"github.com/icholy/digest"
)

type Option func(*callOptions)

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
					InsecureSkipVerify: true, // #nosec G402 - User explicitly requested insecure mode
				}
			}
		}
	}
}

// WithFiles includes file paths in the request.
//
// Deprecated: Use WithMultipartForm() instead. WithMultipartForm() provides:
//   - Support for in-memory file content (not just file paths on disk)
//   - Custom content types (not just application/octet-stream)
//   - Clean separation of files and form fields
//
// WithFiles is kept for backward compatibility only.
//
// Migration example:
//
//	// Before (old way)
//	clientmanager.WithFiles(map[string]string{
//	    "file": "/path/to/image.png",
//	})
//
//	// After (new way)
//	content, _ := os.ReadFile("/path/to/image.png")
//	clientmanager.WithMultipartForm(clientmanager.MultipartForm{
//	    Files: map[string]clientmanager.FilePart{
//	        "file": {
//	            Filename:    "image.png",
//	            Content:     content,
//	            ContentType: "image/png",
//	        },
//	    },
//	})
func WithFiles(files map[string]string) Option {
	return func(co *callOptions) {
		co.files = files
	}
}

// FilePart represents a file in multipart form data with metadata.
//
// The Filename field is used in the Content-Disposition header.
// The Content field contains the raw file content as bytes.
// The ContentType field specifies the MIME type (e.g., "image/png", "application/pdf").
type FilePart struct {
	Filename    string // Name of the file (used in Content-Disposition header)
	Content     []byte // Raw file content
	ContentType string // MIME type (e.g., "image/png", "application/pdf")
}

// MultipartForm represents a complete multipart form with both files and values.
//
// This structure mirrors Go's multipart.Form structure, separating:
//   - Files: Form fields with file data and metadata
//   - Values: Simple string form fields
//
// Example usage:
//
//	clientmanager.WithMultipartForm(clientmanager.MultipartForm{
//	    Files: map[string]clientmanager.FilePart{
//	        "file": {
//	            Filename:    "logo.png",
//	            Content:     imageBytes,
//	            ContentType: "image/png",
//	        },
//	    },
//	    Values: map[string]string{
//	        "alt":      "project logo",
//	        "category": "images",
//	    },
//	})
type MultipartForm struct {
	Files  map[string]FilePart // Field name -> File data with metadata
	Values map[string]string   // Field name -> String value
}

// WithMultipartForm includes multipart form data with both files and string values.
//
// Use this option when you need to upload files with custom content types
// and/or include additional form fields in the same request.
//
// This is the recommended approach for multipart form data as it provides:
//   - Support for in-memory file content (not just file paths on disk)
//   - Custom content types per file (not just application/octet-stream)
//   - Clean separation between files and form fields
//   - Type-safe API for both files and values
//
// Example:
//
//	// Upload file with metadata
//	clientmanager.WithMultipartForm(clientmanager.MultipartForm{
//	    Files: map[string]clientmanager.FilePart{
//	        "file": {
//	            Filename:    "logo.png",
//	            Content:     imageBytes,
//	            ContentType: "image/png",
//	        },
//	    },
//	    Values: map[string]string{
//	        "alt":        "project logo",
//	        "path":       "project_logos",
//	        "filename":   "logo.png",
//	    },
//	})
//
//	// Multiple files with different content types
//	clientmanager.WithMultipartForm(clientmanager.MultipartForm{
//	    Files: map[string]clientmanager.FilePart{
//	        "thumbnail": {
//	            Filename:    "thumb.jpg",
//	            Content:     thumbnailBytes,
//	            ContentType: "image/jpeg",
//	        },
//	        "document": {
//	            Filename:    "report.pdf",
//	            Content:     pdfBytes,
//	            ContentType: "application/pdf",
//	        },
//	    },
//	    Values: map[string]string{
//	        "title":       "Q4 Report",
//	        "description": "Quarterly financial report",
//	    },
//	})
func WithMultipartForm(form MultipartForm) Option {
	return func(co *callOptions) {
		co.multipartForm = form
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
					MinVersion:   tls.VersionTLS12, // #nosec G402 - TLS 1.2+ required
				}
			}
		}
	}
}

func WithRootCertificate(rootCertificate *x509.CertPool) Option {
	return func(co *callOptions) {
		if _, ok := co.client.Transport.(*http.Transport); ok {
			if co.client.Transport.(*http.Transport).TLSClientConfig != nil {
				co.client.Transport.(*http.Transport).TLSClientConfig.RootCAs = rootCertificate
			} else {
				co.client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
					RootCAs:    rootCertificate,
					MinVersion: tls.VersionTLS12, // #nosec G402 - TLS 1.2+ required
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

func WithDisabledHTTP2() Option {
	return func(co *callOptions) {
		if tr, ok := co.client.Transport.(*http.Transport); ok {
			tr.ForceAttemptHTTP2 = false
			tr.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
		}
	}
}
