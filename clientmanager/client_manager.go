package clientmanager

import (
	"context"
	"net/http"
)

type ClientManager[Response any] struct {
	callOptions callOptions
}

func (c ClientManager[Response]) Call(ctx context.Context, endpoint string, options ...Option) (*BaseResponse[Response], error) {
	c.callOptions.setOptions(options...)

	return call[Response](ctx, endpoint, c.callOptions)
}

func (c ClientManager[Response]) CallStream(ctx context.Context, endpoint string, options ...Option) (*StreamResponse, error) {
	c.callOptions.setOptions(options...)

	return callStream(ctx, endpoint, c.callOptions)
}

func (c ClientManager[Response]) CallBytes(ctx context.Context, endpoint string, options ...Option) (*BaseResponse[[]byte], error) {
	c.callOptions.setOptions(options...)

	return callBytes(ctx, endpoint, c.callOptions)
}

func New[Response any](options ...Option) ClientManager[Response] {
	var cOptions = callOptions{
		client: newClient(),
		method: http.MethodGet,
	}

	cOptions.setOptions(options...)

	return ClientManager[Response]{cOptions}
}