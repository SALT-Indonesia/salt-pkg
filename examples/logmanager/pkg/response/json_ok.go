package response

import (
	"encoding/json"
	"net/http"
)

type jsonResponseOk struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func ResponsesWithOK(w http.ResponseWriter, payload interface{}) {
	response, _ := json.Marshal(jsonResponseOk{
		Message: "OK",
		Data:    payload,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)
}
