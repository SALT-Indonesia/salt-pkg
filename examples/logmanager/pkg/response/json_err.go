package response

import (
	"encoding/json"
	"net/http"
)

type jsonResponseErr struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

func ResponsesWithError(w http.ResponseWriter, statusCode int, err error) {
	response, _ := json.Marshal(jsonResponseErr{
		Message: err.Error(),
		Errors: []string{
			err.Error(),
		},
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(response)
}
