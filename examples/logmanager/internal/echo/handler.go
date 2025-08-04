package echo

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	Api SimpleApi
}

func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	sortBy := r.URL.Query().Get("sort")
	nameFilter := r.URL.Query().Get("name")

	response, err := h.Api.Get(r.Context(), map[string]string{
		"sort": sortBy,
		"name": nameFilter,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
