package testdata

import "github.com/SALT-Indonesia/salt-pkg/logmanager"

func NewTx(traceID, name string) *logmanager.Transaction {
	app := logmanager.NewApplication()
	return app.StartHttp(traceID, name)
}

func NewTxRecord(traceID, name string) *logmanager.TxnRecord {
	tx := NewTx(traceID, name)
	return tx.AddDatabase("db")
}
