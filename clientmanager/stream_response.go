package clientmanager

import (
	"io"
	"net/http"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
)

type StreamResponse struct {
	StatusCode int
	Header     http.Header
	Body       io.ReadCloser

	txn *logmanager.TxnRecord
}

func (s *StreamResponse) Close() error {
	var err error
	if s.txn != nil {
		s.txn.End()
		s.txn = nil
	}
	if s.Body != nil {
		err = s.Body.Close()
		s.Body = nil
	}
	return err
}

func (s *StreamResponse) IsSuccess() bool {
	return s.StatusCode >= 200 && s.StatusCode < 300
}