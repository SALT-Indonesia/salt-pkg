package clientmanager

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
)

func isStringType[T any]() bool {
	var t T
	_, ok := any(t).(string)
	return ok
}

func getResponseBody[Response any](raw []byte, contentType string) (Response, error) {
	var response Response
	if len(raw) > 0 {
		switch {
		case strings.Contains(contentType, "application/xml") || strings.Contains(contentType, "text/xml"):
			if err := xml.Unmarshal(raw, &response); err != nil {
				return response, err
			}
		case isStringType[Response]():
			response = any(string(raw)).(Response)
		default:
			if err := json.Unmarshal(raw, &response); err != nil {
				return response, err
			}
		}
	}
	return response, nil
}

func call[Response any](
	ctx context.Context,
	endpoint string,
	cOptions callOptions,
) (*BaseResponse[Response], error) {
	if err := cOptions.validate(); err != nil {
		return nil, err
	}

	req, err := cOptions.getRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	txn := logmanager.StartApiSegment(logmanager.ApiSegment{
		Request: req,
	})
	if txn == nil {
		return nil, errors.New("transaction from the request context cannot be empty")
	}
	defer txn.End()

	res, err := cOptions.client.Do(req)
	if err != nil {
		txn.NoticeError(err)

		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	txn.SetResponse(res)

	raw, _ := io.ReadAll(res.Body)

	response, err := getResponseBody[Response](raw, res.Header.Get("Content-Type"))
	if err != nil {
		txn.NoticeError(err)

		return nil, err
	}

	return &BaseResponse[Response]{
		StatusCode: res.StatusCode,
		Body:       response,
		Raw:        raw,
	}, nil
}

func Call[Response any](ctx context.Context, endpoint string, options ...Option) (*BaseResponse[Response], error) {
	var cOptions = callOptions{
		client: client,
		method: http.MethodGet,
	}

	cOptions.setOptions(options...)

	return call[Response](ctx, endpoint, cOptions)
}
