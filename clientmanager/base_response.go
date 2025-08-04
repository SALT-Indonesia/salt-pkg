package clientmanager

type BaseResponse[T any] struct {
	StatusCode int
	Body       T
	Raw        []byte
}

func (r *BaseResponse[T]) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}
