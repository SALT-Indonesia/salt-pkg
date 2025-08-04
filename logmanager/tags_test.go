package logmanager_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal/test/testdata"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddTags(t *testing.T) {
	tests := []struct {
		name      string
		txnRecord *logmanager.TxnRecord
		inputTags []string
		wantNil   bool
	}{
		{
			name:      "Add multiple tags successfully",
			txnRecord: testdata.NewTxRecord("a", "b"),
			inputTags: []string{"tag1", "tag2"},
			wantNil:   false,
		},
		{
			name:      "Add empty tags",
			txnRecord: testdata.NewTxRecord("a", "b"),
			inputTags: []string{},
			wantNil:   false,
		},
		{
			name:      "Nil txnRecord",
			txnRecord: nil,
			inputTags: []string{"tag1", "tag2"},
			wantNil:   true,
		},
		{
			name:      "Nil tags",
			txnRecord: testdata.NewTxRecord("a", "b"),
			inputTags: nil,
			wantNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.txnRecord.AddTags(tt.inputTags...)
			tt.txnRecord.End()

			if tt.wantNil {
				assert.Nil(t, tt.txnRecord)
				return
			}
			assert.NotNil(t, tt.txnRecord)
		})
	}
}
