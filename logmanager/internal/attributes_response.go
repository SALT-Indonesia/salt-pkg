package internal

// ResponseBodyAttribute adds the response body to the provided Attributes instance using a specified key identifier.
func ResponseBodyAttribute(a *Attributes, bodyBytes []byte) {
	if nil == a || nil == a.value || nil == bodyBytes || len(bodyBytes) == 0 {
		return
	}
	a.value.Add(AttributeResponseBody, toObj(bodyBytes))
}
