package logmanager_test

import (
	"context"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal/test/testdata"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContext(t *testing.T) {
	type args struct {
		ctx context.Context
		txn *logmanager.Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
	}{
		{
			name: "nil transaction",
			args: args{
				ctx: context.Background(),
				txn: nil,
			},
			wantNil: true,
		},
		{
			name: "valid transaction",
			args: args{
				ctx: context.Background(),
				txn: testdata.NewTx("123", "name"),
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := logmanager.NewContext(tt.args.ctx, tt.args.txn)
			tx := logmanager.FromContext(ctx)
			if tt.wantNil {
				assert.Nil(t, tx)
				return
			}
			assert.NotNil(t, tx)
		})
	}
}

func TestFromContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantNil bool
	}{
		{
			name:    "nil context",
			ctx:     nil,
			wantNil: true,
		},
		{
			name:    "empty context",
			ctx:     context.Background(),
			wantNil: true,
		},
		{
			name: "context with transaction",
			ctx: func() context.Context {
				txn := testdata.NewTx("123", "name")
				return logmanager.NewContext(context.Background(), txn)
			}(),
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn := logmanager.FromContext(tt.ctx)
			if tt.wantNil {
				assert.Nil(t, txn)
			} else {
				assert.NotNil(t, txn)
			}
		})
	}
}

func TestRequestWithTransactionContext(t *testing.T) {
	tests := []struct {
		name    string
		req     *http.Request
		txn     *logmanager.Transaction
		wantNil bool
	}{
		{
			name:    "nil transaction",
			req:     &http.Request{},
			txn:     nil,
			wantNil: true,
		},
		{
			name:    "valid transaction",
			req:     &http.Request{},
			txn:     testdata.NewTx("123", "name"),
			wantNil: false,
		},
		{
			name:    "nil request",
			req:     nil,
			txn:     testdata.NewTx("123", "name"),
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := logmanager.RequestWithTransactionContext(tt.req, tt.txn)
			if tt.req == nil {
				assert.Nil(t, req)
				return
			}
			txn := logmanager.FromContext(req.Context())
			if tt.wantNil {
				assert.Nil(t, txn)
			} else {
				assert.NotNil(t, txn)
			}
		})
	}
}

func TestRequestWithContext(t *testing.T) {
	tests := []struct {
		name    string
		req     *http.Request
		key     logmanager.ContextKey
		value   string
		wantNil bool
	}{
		{
			name:    "empty key and value",
			req:     &http.Request{},
			key:     "",
			value:   "",
			wantNil: true,
		},
		{
			name:    "empty key",
			req:     &http.Request{},
			key:     "",
			value:   "value",
			wantNil: true,
		},
		{
			name:    "empty value",
			req:     &http.Request{},
			key:     "key",
			value:   "",
			wantNil: true,
		},
		{
			name:    "valid key and value",
			req:     &http.Request{},
			key:     "key",
			value:   "value",
			wantNil: false,
		},
		{
			name:    "nil request",
			req:     nil,
			key:     "key",
			value:   "value",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := logmanager.RequestWithContext(tt.req, tt.key, tt.value)
			if tt.req == nil || tt.key == "" || tt.value == "" {
				assert.Equal(t, tt.req, req)
				return
			}
			v := req.Context().Value(tt.key)
			if tt.wantNil {
				assert.Nil(t, v)
			} else {
				assert.Equal(t, tt.value, v)
			}
		})
	}
}

func TestCloneTransactionToContext(t *testing.T) {
	tests := []struct {
		name    string
		srcCtx  context.Context
		dstCtx  context.Context
		wantNil bool
	}{
		{
			name:    "nil source context",
			srcCtx:  nil,
			dstCtx:  context.Background(),
			wantNil: true,
		},
		{
			name:    "nil destination context",
			srcCtx:  context.Background(),
			dstCtx:  nil,
			wantNil: true,
		},
		{
			name:    "source context without transaction",
			srcCtx:  context.Background(),
			dstCtx:  context.Background(),
			wantNil: true,
		},
		{
			name: "source context with transaction",
			srcCtx: func() context.Context {
				txn := testdata.NewTx("123", "name")
				return logmanager.NewContext(context.Background(), txn)
			}(),
			dstCtx:  context.Background(),
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := logmanager.CloneTransactionToContext(tt.srcCtx, tt.dstCtx)
			if tt.srcCtx == nil || tt.dstCtx == nil {
				assert.Equal(t, tt.dstCtx, ctx)
				return
			}
			txn := logmanager.FromContext(ctx)
			if tt.wantNil {
				assert.Nil(t, txn)
			} else {
				assert.NotNil(t, txn)
				// Verify that the transaction in the destination context is the same as the one in the source context
				srcTxn := logmanager.FromContext(tt.srcCtx)
				assert.Equal(t, srcTxn, txn)
			}
		})
	}
}
