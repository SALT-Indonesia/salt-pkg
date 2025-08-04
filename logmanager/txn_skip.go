package logmanager

// SkipResponse sets the transaction record to skip the processing or logging of the response. Does nothing if the record is nil.
func (txn *TxnRecord) SkipResponse() {
	if txn == nil {
		return
	}
	txn.skipResponse = true
}

// SkipRequest sets the transaction record to skip the processing or logging of the request. Does nothing if the record is nil.
func (txn *TxnRecord) SkipRequest() {
	if txn == nil {
		return
	}
	txn.skipRequest = true
}
