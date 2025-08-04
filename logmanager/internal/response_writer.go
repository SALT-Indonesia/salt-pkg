package internal

import "net/http"

type DummyResponseWriter struct{}

func (rw DummyResponseWriter) Header() http.Header { return nil }

func (rw DummyResponseWriter) Write(_ []byte) (int, error) { return 0, nil }

func (rw DummyResponseWriter) WriteHeader(_ int) {}
