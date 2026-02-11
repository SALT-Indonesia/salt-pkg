package logmanager

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"net/http"
	"net/url"
)

type WebRequest struct {
	Header http.Header
	URL    *url.URL
	Method string
	Host   string
}

type WebResponse struct {
	StatusCode int
	Body       []byte
}

// SetResponse updates the transaction record with the response body and status code attributes.
// If the TxnRecord is nil, the function simply returns without making changes.
func (txn *TxnRecord) SetResponse(resp *http.Response) {
	if nil == txn || nil == resp {
		return
	}
	internal.ResponseBodyAttributes(txn.attrs, resp)
	internal.ResponseCodeAttribute(txn.attrs, resp.StatusCode)
}

func (txn *TxnRecord) SetResponseBodyAndCode(body []byte, code int) {
	if nil == txn || nil == body {
		return
	}
	internal.ResponseBodyAttribute(txn.attrs, body)
	internal.ResponseCodeAttribute(txn.attrs, code)
}

// SetWebRequest sets the web request details for the transaction record.
// It extracts headers, URL, method, and host from the provided http.Request
// and updates the transaction with these Attributes. If the request is nil,
// the transaction is updated with an empty WebRequest object.
func (txn *TxnRecord) SetWebRequest(r *http.Request) {
	if nil == txn {
		return
	}
	if nil == r {
		txn.setWebRequest(WebRequest{})
		return
	}
	wr := WebRequest{
		Header: r.Header,
		URL:    r.URL,
		Method: r.Method,
		Host:   r.Host,
	}
	txn.setWebRequest(wr)
	internal.RequestBodyAttributes(txn.attrs, r)
}

// SetWebRequestRaw sets raw web request details, including method, headers, URL, and host, along with the request body.
func (txn *TxnRecord) SetWebRequestRaw(body interface{}, r WebRequest) {
	if nil == txn {
		return
	}

	h := r.Header
	internal.RequestAgentAttributes(txn.attrs, r.Method, h, r.URL, r.Host)
	internal.RequestBodyAttribute(txn.attrs, body)
}

func (txn *TxnRecord) setWebRequest(r WebRequest) {
	h := r.Header
	internal.RequestAgentAttributes(txn.attrs, r.Method, h, r.URL, r.Host)
}

// CaptureMultipartFormData captures multipart form data from the request after
// the handler has parsed it. This should be called after the handler runs to capture
// form fields and file metadata that were parsed during request processing.
// It only captures data if the request has multipart/form-data content type and
// the form has been parsed (r.MultipartForm is not nil). If request body was already
// captured (e.g., for JSON requests), this method does nothing.
func (txn *TxnRecord) CaptureMultipartFormData(r *http.Request) {
	if nil == txn || nil == r {
		return
	}
	internal.CaptureMultipartFormDataIfParsed(txn.attrs, r)
}

// SetWebResponseHttp wraps the given http.ResponseWriter with a replacementResponseWriter that tracks transaction details.
func (txn *TxnRecord) SetWebResponseHttp(w http.ResponseWriter) http.ResponseWriter {
	if nil == txn || w == nil {
		w = internal.DummyResponseWriter{}
	}

	return &replacementResponseWriter{
		thd:      txn,
		original: w,
	}
}

// SetWebResponse sets the web response details for a transaction record.
// It updates the transaction's Attributes with the response body and status code.
// If the transaction record is nil, this method does nothing.
func (txn *TxnRecord) SetWebResponse(w WebResponse) {
	if nil == txn {
		return
	}

	internal.ResponseBodyAttribute(txn.attrs, w.Body)
	internal.ResponseCodeAttribute(txn.attrs, w.StatusCode)
}
