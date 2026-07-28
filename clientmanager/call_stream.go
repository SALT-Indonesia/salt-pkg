package clientmanager

import (
	"context"
	"io"
	"net/http"
)

func callStream(ctx context.Context, endpoint string, cOptions callOptions) (*StreamResponse, error) {
	res, txn, err := execute(ctx, endpoint, cOptions)
	if err != nil {
		return nil, err
	}

	return &StreamResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header.Clone(),
		Body:       res.Body,
		txn:        txn,
	}, nil
}

func callBytes(ctx context.Context, endpoint string, cOptions callOptions) (*BaseResponse[[]byte], error) {
	res, txn, err := execute(ctx, endpoint, cOptions)
	if err != nil {
		return nil, err
	}
	defer txn.End()
	defer func() {
		_ = res.Body.Close()
	}()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		txn.NoticeError(err)

		return nil, err
	}

	return &BaseResponse[[]byte]{
		StatusCode: res.StatusCode,
		Body:       raw,
		Raw:        raw,
	}, nil
}

func CallStream(ctx context.Context, endpoint string, options ...Option) (*StreamResponse, error) {
	var cOptions = callOptions{
		client: client,
		method: http.MethodGet,
	}

	cOptions.setOptions(options...)

	return callStream(ctx, endpoint, cOptions)
}

func CallBytes(ctx context.Context, endpoint string, options ...Option) (*BaseResponse[[]byte], error) {
	var cOptions = callOptions{
		client: client,
		method: http.MethodGet,
	}

	cOptions.setOptions(options...)

	return callBytes(ctx, endpoint, cOptions)
}