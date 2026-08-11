package clientmanager

import "net/http"

type BaseResponse[T any] struct {
	StatusCode int
	Header     http.Header
	Body       T
	Raw        []byte
}

func (r *BaseResponse[T]) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}
