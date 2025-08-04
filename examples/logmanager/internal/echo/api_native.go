package echo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"net/http"
	"net/url"
)

type APINative struct {
	client  *http.Client
	baseURL string
}

func (a *APINative) Get(ctx context.Context, params map[string]string) (*Response, error) {
	u, err := url.Parse(a.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	q := u.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	txn := logmanager.StartApiSegment(logmanager.ApiSegment{Request: req}) // start here
	defer txn.End()                                                        // end here

	resp, err := a.client.Do(req)
	if err != nil {
		txn.NoticeError(err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	txn.SetResponse(resp) // remember to save the response here

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		txn.NoticeError(err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func NewAPINative(baseURL string) *APINative {
	return &APINative{
		client:  http.DefaultClient,
		baseURL: baseURL,
	}
}
