package logmanager

import (
	"context"
	"net/http"
)

type ContextKey string

const (
	TransactionContextKey ContextKey = "transaction"
	TraceIDContextKey     ContextKey = "trace_id"
)

func (c ContextKey) String() string {
	return string(c)
}

// ToContext returns a new context with the current Transaction embedded, allowing it to be retrieved later.
func (t *Transaction) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, TransactionContextKey.String(), t)
}

// FromContext retrieves a Transaction from the given context if present.
// It returns nil if the context is nil or if no Transaction is associated with the context.
func FromContext(ctx context.Context) *Transaction {
	if nil == ctx {
		return nil
	}

	if val, ok := ctx.Value(TransactionContextKey.String()).(*Transaction); ok {
		return val
	}
	return nil
}

func transactionFromRequestContext(req *http.Request) *Transaction {
	var tx *Transaction
	if nil != req {
		tx = FromContext(req.Context())
	}
	return tx
}

// NewContext returns a new context.Context that carries a Transaction instance using a specific context key.
func NewContext(ctx context.Context, txn *Transaction) context.Context {
	if nil == txn {
		return ctx
	}
	return context.WithValue(ctx, TransactionContextKey.String(), txn)
}

// RequestWithTransactionContext attaches a Transaction to the given HTTP request's context.
// If the txn parameter is nil, it returns the request unchanged.
// Otherwise, it creates a new context containing the transaction and returns a new request with the updated context.
func RequestWithTransactionContext(req *http.Request, txn *Transaction) *http.Request {
	if nil == txn || nil == req {
		return req
	}

	ctx := req.Context()
	ctx = NewContext(ctx, txn)
	return req.WithContext(ctx)
}

// RequestWithContext attaches a key-value pair to the HTTP request's context and returns the modified request.
func RequestWithContext(req *http.Request, key ContextKey, value string) *http.Request {
	if "" == key || "" == value || nil == req {
		return req
	}

	ctx := req.Context()
	ctx = context.WithValue(ctx, key, value)
	return req.WithContext(ctx)
}

// CloneTransactionToContext clones a transaction from a source context to a destination context.
// This is useful for goroutines that need to maintain the same transaction information but in a different context.
// If no transaction is found in the source context, the destination context is returned unchanged.
func CloneTransactionToContext(srcCtx, dstCtx context.Context) context.Context {
	if srcCtx == nil || dstCtx == nil {
		return dstCtx
	}

	txn := FromContext(srcCtx)
	if txn == nil {
		return dstCtx
	}

	return NewContext(dstCtx, txn)
}
