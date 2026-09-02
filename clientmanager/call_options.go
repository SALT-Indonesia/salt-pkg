package clientmanager

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"syscall"
	"time"

	"github.com/Azure/go-ntlmssp"
	"github.com/dghubble/oauth1"
	"github.com/icholy/digest"
	validator "github.com/go-playground/validator/v10"
	"golang.org/x/oauth2"
)

var (
	validate     *validator.Validate
	validateOnce sync.Once
)

type callOptions struct {
	client                *http.Client
	auth                  Auth
	host                  string
	headers               http.Header
	method                string
	isFormURLEncoded      bool
	files                 map[string]string   // Keep for backward compatibility (deprecated)
	multipartForm         MultipartForm        // Enhanced multipart support
	requestBody           any
	urlValues             url.Values
	bodyReader            io.Reader
	bodyReaderContentType string
	dialerControl         func(network, address string, c syscall.RawConn) error
	dialTimeout           time.Duration
	dialKeepAlive         time.Duration
	maxResponseBytes      int64
}

func (c *callOptions) setOptions(options ...Option) {
	for _, option := range options {
		option(c)
	}
	c.resolve()
}

// resolve applies deferred settings that depend on the combination of options.
// Currently it rebuilds the transport dialer when a Control function is set.
func (c *callOptions) resolve() {
	if c.dialerControl == nil {
		return
	}
	tr := c.baseTransport()
	if tr == nil {
		return
	}
	timeout := c.dialTimeout
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	keepAlive := c.dialKeepAlive
	if keepAlive <= 0 {
		keepAlive = 60 * time.Second
	}
	tr.DialContext = (&net.Dialer{
		Timeout:   timeout,
		KeepAlive: keepAlive,
		Control:   c.dialerControl,
	}).DialContext
}

// baseTransport walks through transport wrappers (digest, NTLM, OAuth1/2) to
// find the underlying *http.Transport that owns the dialer.
func (c *callOptions) baseTransport() *http.Transport {
	tr := c.client.Transport
	for {
		switch t := tr.(type) {
		case *http.Transport:
			return t
		case *digest.Transport:
			tr = t.Transport
			continue
		case ntlmssp.Negotiator:
			tr = t.RoundTripper
			continue
		case *oauth1.Transport:
			tr = t.Base
			continue
		case *oauth2.Transport:
			tr = t.Base
			continue
		}
		return nil
	}
}

func (c callOptions) validate() error {
	validateOnce.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())
	})
	if c.requestBody != nil {
		if err := validate.Struct(c.requestBody); err != nil {
			var invalidValidationError *validator.InvalidValidationError
			if !errors.As(err, &invalidValidationError) {
				return err
			}
		}
	}
	return nil
}

func (c callOptions) getRequestBody() (io.Reader, string, error) {
	switch {
	case c.bodyReader != nil:
		return c.bodyReader, c.bodyReaderContentType, nil
	case len(c.multipartForm.Files) > 0 || len(c.multipartForm.Values) > 0:
		body, contentType := getMultipartFormBody(c.multipartForm)

		return body, contentType, nil
	case len(c.files) > 0:
		body, contentType, err := getFilesBody(c.files, c.requestBody)
		if err != nil {
			return nil, "", err
		}

		return body, contentType, nil
	case c.isFormURLEncoded:
		body, contentType := getFormURLEncodedBody(c.requestBody)

		return body, contentType, nil
	default:
		body := getJSONBody(c.requestBody)
		contentType := "application/json"
		var reqBody io.Reader
		if body != nil {
			reqBody = body
		}

		return reqBody, contentType, nil
	}
}

func (c callOptions) addURLValues() string {
	if c.urlValues != nil {
		return "?" + c.urlValues.Encode()
	}
	return ""
}

func (c callOptions) setRequestHeaders(req *http.Request, contentType string) error {
	if c.headers != nil {
		req.Header = c.headers
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.auth != nil {
		if err := c.auth(req); err != nil {
			return err
		}
	}
	return nil
}

func (c callOptions) getRequest(ctx context.Context, endpoint string) (*http.Request, error) {
	body, contentType, err := c.getRequestBody()
	if err != nil {
		return nil, err
	}

	endpoint += c.addURLValues()
	req, err := http.NewRequestWithContext(ctx, c.method, c.host+endpoint, body)
	if err != nil {
		return nil, err
	}

	if err := c.setRequestHeaders(req, contentType); err != nil {
		return nil, err
	}
	return req, nil
}