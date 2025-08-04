package testdata

import (
	"bytes"
	"context"
	"net/http"
)

func NewRequestWithCtx() *http.Request {
	tx := NewTx("1234567890", "sample")
	req, _ := http.NewRequestWithContext(tx.ToContext(context.TODO()), http.MethodPost, "/sample", bytes.NewBuffer([]byte(`{"name":"product"}`)))
	return req
}

func NewRequestWithEmptyCtx() *http.Request {
	req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, "/sample", bytes.NewBuffer([]byte(`{"name":"product"}`)))
	return req
}
