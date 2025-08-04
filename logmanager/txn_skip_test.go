package logmanager_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"testing"
)

func TestTxnRecord_SkipResponse(t *testing.T) {
	t.Run("it should not nothing with nil txn", func(t *testing.T) {
		txn := (*logmanager.TxnRecord)(nil)
		txn.SkipResponse()
	})

	t.Run("it should not print log response body with skip response", func(t *testing.T) {
		txn := &logmanager.TxnRecord{}
		txn.SkipResponse()
	})
}

func TestTxnRecord_SkipRequest(t *testing.T) {
	t.Run("it should not nothing with nil txn", func(t *testing.T) {
		txn := (*logmanager.TxnRecord)(nil)
		txn.SkipRequest()
	})

	t.Run("it should not print log request body with skip request", func(t *testing.T) {
		txn := &logmanager.TxnRecord{}
		txn.SkipRequest()
	})
}
