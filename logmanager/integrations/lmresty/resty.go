package lmresty

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"github.com/go-resty/resty/v2"
	"net/http"
)

func NewTxn(resp *resty.Response) *logmanager.TxnRecord {
	if nil == resp {
		return nil
	}

	tx := logmanager.FromContext(resp.Request.Context())
	if nil == tx {
		return nil
	}

	txn := tx.AddTxnNow(internal.GetCaller(), logmanager.TxnTypeApi, resp.Request.Time)

	reqBody := resp.Request.Body
	if resp.Request.Method == http.MethodGet {
		reqBody = resp.Request.QueryParam
	}

	txn.SetWebRequestRaw(reqBody, logmanager.WebRequest{
		Header: resp.Request.RawRequest.Header,
		URL:    resp.Request.RawRequest.URL,
		Method: resp.Request.RawRequest.Method,
		Host:   resp.Request.RawRequest.Host,
	})

	txn.SetResponseBodyAndCode(resp.Body(), resp.StatusCode())
	return txn
}

// NewTxnWithMasking creates a new transaction record with combined masking applied to request and response data.
// This combines go-masker struct tags with logmanager's advanced masking configurations.
func NewTxnWithMasking(resp *resty.Response, maskingConfigs []logmanager.MaskingConfig) *logmanager.TxnRecord {
	if nil == resp {
		return nil
	}

	tx := logmanager.FromContext(resp.Request.Context())
	if nil == tx {
		return nil
	}

	txn := tx.AddTxnNow(internal.GetCaller(), logmanager.TxnTypeApi, resp.Request.Time)

	reqBody := resp.Request.Body
	if resp.Request.Method == http.MethodGet {
		reqBody = resp.Request.QueryParam
	}

	// Apply combined masking to request body (struct tags + configurations)
	if reqBody != nil {
		if maskedReqBody, err := logmanager.StructMaskWithConfig(reqBody, maskingConfigs); err == nil {
			reqBody = maskedReqBody
		}
	}

	txn.SetWebRequestRaw(reqBody, logmanager.WebRequest{
		Header: resp.Request.RawRequest.Header,
		URL:    resp.Request.RawRequest.URL,
		Method: resp.Request.RawRequest.Method,
		Host:   resp.Request.RawRequest.Host,
	})

	// Apply combined masking to response body (struct tags + configurations)
	respBody := resp.Body()
	if maskedRespBody, err := logmanager.StructMaskWithConfig(respBody, maskingConfigs); err == nil {
		if maskedBytes, ok := maskedRespBody.([]byte); ok {
			respBody = maskedBytes
		}
	}

	txn.SetResponseBodyAndCode(respBody, resp.StatusCode())
	return txn
}

// NewTxnWithConfig creates a new transaction record using the provided masking configurations.
// This is the recommended method for masking support. Masking is always applied when using this function.
func NewTxnWithConfig(resp *resty.Response, maskingConfigs []logmanager.MaskingConfig) *logmanager.TxnRecord {
	return NewTxnWithMasking(resp, maskingConfigs)
}

// Convenience functions for common masking patterns

// NewTxnWithPasswordMasking creates a transaction with common password field masking
func NewTxnWithPasswordMasking(resp *resty.Response) *logmanager.TxnRecord {
	configs := []logmanager.MaskingConfig{
		{
			FieldPattern: "password",
			Type:         logmanager.FullMask,
		},
		{
			FieldPattern: "token",
			Type:         logmanager.FullMask,
		},
		{
			FieldPattern: "secret",
			Type:         logmanager.FullMask,
		},
		{
			JSONPath: "$.authorization",
			Type:     logmanager.FullMask,
		},
	}
	return NewTxnWithMasking(resp, configs)
}

// NewTxnWithEmailMasking creates a transaction with email masking
// Masks email addresses preserving domain and showing first/last chars of username
// Example: arfan.azhari@salt.id → ar******ri@salt.id
func NewTxnWithEmailMasking(resp *resty.Response) *logmanager.TxnRecord {
	configs := []logmanager.MaskingConfig{
		{
			FieldPattern: "email",
			Type:         logmanager.EmailMask,
			ShowFirst:    2, // Show first 2 chars of username
			ShowLast:     2, // Show last 2 chars of username
		},
	}
	return NewTxnWithMasking(resp, configs)
}

// NewTxnWithCreditCardMasking creates a transaction with credit card number masking
func NewTxnWithCreditCardMasking(resp *resty.Response) *logmanager.TxnRecord {
	configs := []logmanager.MaskingConfig{
		{
			FieldPattern: "card",
			Type:         logmanager.PartialMask,
			ShowFirst:    4,
			ShowLast:     4,
		},
		{
			FieldPattern: "cvv",
			Type:         logmanager.FullMask,
		},
		{
			FieldPattern: "pan",
			Type:         logmanager.PartialMask,
			ShowFirst:    4,
			ShowLast:     4,
		},
	}
	return NewTxnWithMasking(resp, configs)
}
