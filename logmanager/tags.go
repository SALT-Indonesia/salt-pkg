package logmanager

func (txn *TxnRecord) AddTags(tags ...string) {
	if txn == nil || txn.attrs == nil {
		return
	}

	for _, tag := range tags {
		if tag != "" {
			txn.tags = append(txn.tags, tag)
		}
	}
}
