package echo

import "context"

type Response struct {
	Headers struct {
		Host         string `json:"host"`
		Connection   string `json:"connection"`
		XAmznTraceId string `json:"x-amzn-trace-id"`
		Accept       string `json:"accept"`
	} `json:"headers"`
	Url string `json:"url"`
}

type SimpleApi interface {
	Get(ctx context.Context, params map[string]string) (*Response, error)
}
