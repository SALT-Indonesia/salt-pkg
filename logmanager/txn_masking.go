package logmanager

import (
	"encoding/json"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
)

// SetResponseBodyAndCodeMasked updates the transaction attributes with the masked response body and HTTP status code.
// It applies masking configurations to the response body before logging.
func (txn *TxnRecord) SetResponseBodyAndCodeMasked(body []byte, code int, maskingConfigs []MaskingConfig) {
	if nil == txn || nil == body {
		return
	}

	// Apply masking to response body
	maskedBody := body
	if len(maskingConfigs) > 0 {
		internalConfigs := ConvertMaskingConfigs(maskingConfigs)
		jsonMasker := internal.NewJSONMasker(internalConfigs)
		if parsed, err := parseJSON(body); err == nil {
			maskedData := jsonMasker.MaskData(parsed)
			if maskedBytes, err := marshalJSON(maskedData); err == nil {
				maskedBody = maskedBytes
			}
		}
	}

	internal.ResponseBodyAttribute(txn.attrs, maskedBody)
	internal.ResponseCodeAttribute(txn.attrs, code)
}

// SetWebRequestRawMasked sets raw web request details with masking applied to the request body.
// It applies masking configurations to the request body before logging.
func (txn *TxnRecord) SetWebRequestRawMasked(body interface{}, r WebRequest, maskingConfigs []MaskingConfig) {
	if nil == txn {
		return
	}

	h := r.Header
	internal.RequestAgentAttributes(txn.attrs, r.Method, h, r.URL, r.Host)

	// Apply masking to request body
	maskedBody := body
	if len(maskingConfigs) > 0 && body != nil {
		internalConfigs := ConvertMaskingConfigs(maskingConfigs)
		jsonMasker := internal.NewJSONMasker(internalConfigs)
		maskedBody = jsonMasker.MaskData(body)
	}

	internal.RequestBodyAttribute(txn.attrs, maskedBody)
}

// Helper functions for JSON parsing and marshaling in masked methods
func parseJSON(data []byte) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func marshalJSON(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}
