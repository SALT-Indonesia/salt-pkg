package internal

func RequestBodyConsumerAttributes(a *Attributes, bodyBytes []byte) {
	if nil == bodyBytes || len(bodyBytes) == 0 {
		return
	}

	a.value.Add(AttributeConsumerRequestBody, toObj(bodyBytes))
}
