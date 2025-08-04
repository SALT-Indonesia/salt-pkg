package logmanager

import (
	"context"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"net/http"
)

type GrpcSegment struct {

	// Url represents the endpoint or path of a gRPC service being accessed.
	Url string

	// Request represents the payload or body of a gRPC request, and can be of any data type.
	Request interface{}
}

// StartGrpcSegment initializes a gRPC transaction segment, adding it to the provided Transaction if not nil.
// It uses the gRPC service's URL and request body for creating the segment.
// Returns a new TxnRecord containing details of the gRPC segment if the transaction is valid, otherwise returns nil.
func StartGrpcSegment(tx *Transaction, i GrpcSegment) *TxnRecord {
	if nil == tx {
		return nil
	}

	txn := tx.AddTxn(i.Url, TxnTypeGrpc)
	txn.attrs.Value().Add(internal.AttributeRequestBody, i.Request)
	return txn
}

func (txn *TxnRecord) SetResponseValue(value interface{}) {
	if nil == txn {
		return
	}

	txn.attrs.Value().Add(internal.AttributeResponseBody, value)
}

// SetRequestValue adds a request body value to the transaction attributes if the transaction record is not nil.
func (txn *TxnRecord) SetRequestValue(value interface{}) {
	if nil == txn {
		return
	}

	txn.attrs.Value().Add(internal.AttributeRequestBody, value)
}

type ApiSegment struct {

	// Name represents the identifier for the API segment.
	// This field can be empty. If it is empty, the method name
	// of the caller will be used as the identifier.
	Name string

	// Request is an HTTP request struct used for initiating transactions in the ApiSegment.
	Request *http.Request
}

// StartApiSegment initializes a new transaction record for an API segment using the provided ApiSegment struct.
// If the transaction is nil, it returns nil. If the ApiSegment's Name is empty, the caller's name is used as the Name.
// It also adds the trace ID from the transaction to the X-Trace-Id header in the HTTP request.
func StartApiSegment(i ApiSegment) *TxnRecord {
	tx := transactionFromRequestContext(i.Request)
	if nil == tx {
		return nil
	}

	if i.Name == "" {
		i.Name = internal.GetCaller()
	}

	if i.Request != nil {
		i.Request.Header.Set(tx.traceIDHeaderKey, tx.traceID)
	}

	txn := tx.AddTxn(i.Name, TxnTypeApi)
	txn.SetWebRequest(i.Request)
	return txn
}

type DatabaseSegment struct {

	// Name represents the identifier for the database segment, often used to denote a specific action or query performed.
	Name string

	// Table specifies the database table associated with the segment, indicating which table the operation interacts with.
	Table string

	// Query is a string representing the database query executed within the corresponding DatabaseSegment.
	Query string

	// Host represents the hostname or address of the database server for the segment.
	Host string
}

// StartDatabaseSegment initiates a database transaction segment with the provided transaction and database segment info.
// If the transaction is nil, the function returns nil. It assigns a default name to the database segment if not provided.
func StartDatabaseSegment(tx *Transaction, i DatabaseSegment) *TxnRecord {
	if nil == tx {
		return nil
	}

	if i.Name == "" {
		i.Name = internal.GetCaller()
	}

	txn := tx.AddTxn(i.Name, TxnTypeDatabase)
	txn.attrs.Value().AddString(internal.AttributeDatabaseTable, i.Table)
	txn.attrs.Value().AddString(internal.AttributeDatabaseQuery, i.Query)
	txn.attrs.Value().AddString(internal.AttributeDatabaseHost, i.Host)
	return txn
}

type OtherSegment struct {
	Name  string
	Extra map[string]interface{}
}

// StartOtherSegment creates a new transaction record for an "other" segment within a transaction.
// It uses the provided OtherSegment information to set the name and extra attributes.
// If the segment's name is empty, it derives the name from the calling function.
// Returns the created TxnRecord or nil if the transaction is nil.
func StartOtherSegment(tx *Transaction, i OtherSegment) *TxnRecord {
	if nil == tx {
		return nil
	}

	if i.Name == "" {
		i.Name = internal.GetCaller()
	}

	txn := tx.AddTxn(i.Name, TxnTypeOther)
	for k, v := range i.Extra {
		txn.attrs.Value().Add(k, v)
	}
	return txn
}

// StartOtherSegmentWithContext creates a new transaction record for an "other" segment within a transaction,
// retrieving the transaction from the provided context.Context.
// It uses the provided OtherSegment information to set the name and extra attributes.
// If the segment's name is empty, it derives the name from the calling function.
// Returns the created TxnRecord or nil if the transaction cannot be retrieved from the context or is nil.
func StartOtherSegmentWithContext(ctx context.Context, i OtherSegment) *TxnRecord {
	tx := FromContext(ctx)
	if nil == tx {
		return nil
	}

	if i.Name == "" {
		i.Name = internal.GetCaller()
	}

	txn := tx.AddTxn(i.Name, TxnTypeOther)
	for k, v := range i.Extra {
		txn.attrs.Value().Add(k, v)
	}
	return txn
}

// StartOtherSegmentWithMessage creates a new transaction record for an "other" segment within a transaction,
// retrieving the transaction from the provided context.Context and using the given message as the segment name.
// Returns the created TxnRecord or nil if the transaction cannot be retrieved from the context or is nil.
func StartOtherSegmentWithMessage(ctx context.Context, message string) {
	tx := FromContext(ctx)
	if nil == tx {
		return
	}

	txn := tx.AddTxn(message, TxnTypeOther)
	defer txn.End()
}
