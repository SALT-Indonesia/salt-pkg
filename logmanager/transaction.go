package logmanager

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"sync"
	"time"
)

type Transaction struct {
	traceID string
	*TxnRecord
	txnRecords       map[string]*TxnRecord
	tags             []string
	traceIDKey       string
	traceIDHeaderKey string
	mu               sync.Mutex
}

// ExposeAllHeaders modifies the transaction's configuration to enable exposing all headers.
// Don't do this on production
func (t *Transaction) ExposeAllHeaders() {
	t.exposeAllHeader = true
}

// newEmptyTransaction creates and returns a new Transaction instance with default initialized attributes.
func newEmptyTransaction() *Transaction {
	return &Transaction{
		TxnRecord: &TxnRecord{
			attrs: internal.NewAttributes(),
		},
	}
}

// AddTxn creates a new transaction record with the specified name and transaction type, adding it to the transaction's records.
func (t *Transaction) AddTxn(name string, logType TxnType) *TxnRecord {
	if nil == t {
		return nil
	}

	return t.AddTxnNow(name, logType, time.Now())
}

// AddTxnNow creates and returns a new transaction record with the specified name, type, and start time.
func (t *Transaction) AddTxnNow(name string, logType TxnType, start time.Time) *TxnRecord {
	if nil == t {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	s := &TxnRecord{
		name:          name,
		traceID:       t.traceID,
		start:         start,
		txnType:       logType,
		attrs:         internal.NewAttributes(),
		service:       t.service,
		logger:        t.logger,
		tags:          t.tags,
		exposeHeaders: t.exposeHeaders,
		debug:         t.debug,
		traceIDKey:    t.traceIDKey,
	}
	t.txnRecords[name] = s
	return s
}

// AddDatabase creates a new database transaction record with the provided name and adds it to the transaction's records.
func (t *Transaction) AddDatabase(name string) *TxnRecord {
	if nil == t {
		return nil
	}

	s := t.AddTxn(name, TxnTypeDatabase)
	t.txnRecords[name] = s
	return s
}

func (t *Transaction) TraceID() string {
	return t.traceID
}
