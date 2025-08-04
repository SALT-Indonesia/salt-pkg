package echo

import (
	"context"
	"fmt"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmresty"
	"github.com/go-resty/resty/v2"
)

type APIResty struct {
	client  *resty.Client
	baseURL string
}

func (a *APIResty) Get(ctx context.Context, params map[string]string) (*Response, error) {
	resp, err := a.client.R().
		SetContext(ctx).
		SetQueryParams(params).
		SetResult(&Response{}).
		Get(fmt.Sprintf("%s/get", a.baseURL))

	if err != nil {
		return nil, err
	}

	txn := lmresty.NewTxn(resp) // start here
	defer txn.End()             // end here

	if resp.IsError() {
		return nil, fmt.Errorf("APIResty call failed with status code: %d", resp.StatusCode())
	}

	return resp.Result().(*Response), nil
}

func NewApiResty(baseURL string) *APIResty {
	return &APIResty{
		client:  resty.New(),
		baseURL: baseURL,
	}
}
