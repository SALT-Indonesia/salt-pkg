package clientmanager

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"

	validator "github.com/go-playground/validator/v10"
)

var validate *validator.Validate

type callOptions struct {
	client           *http.Client
	auth             Auth
	host             string
	headers          http.Header
	method           string
	isFormURLEncoded bool
	files            map[string]string
	requestBody      any
	urlValues        url.Values
}

func (c *callOptions) setOptions(options ...Option) {
	for _, option := range options {
		option(c)
	}
}

func (c callOptions) validate() error {
	if validate == nil {
		validate = validator.New(validator.WithRequiredStructEnabled())
	}
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
	var (
		body        *bytes.Buffer
		contentType string
		err         error
	)
	switch {
	case len(c.files) > 0:
		body, contentType, err = getFilesBody(c.files, c.requestBody)
		if err != nil {
			return nil, "", err
		}
	case c.isFormURLEncoded:
		body, contentType = getFormURLEncodedBody(c.requestBody)
	default:
		body, contentType = getJSONBody(c.requestBody)
	}

	var reqBody io.Reader
	if body != nil {
		reqBody = body
	}
	return reqBody, contentType, err
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
