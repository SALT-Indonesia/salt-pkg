package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgorilla"
	"github.com/gorilla/mux"
)

func main() {
	app := logmanager.NewApplication(
		logmanager.WithAppName("http-gorilla"),
		logmanager.WithMaskingConfig([]logmanager.MaskingConfig{
			{
				Type:     logmanager.FullMask,
				JSONPath: "$..token",
			},
			{
				Type:     logmanager.FullMask,
				JSONPath: "$..password",
			},
			{
				Type:      logmanager.PartialMask,
				JSONPath:  "$..apiKey",
				ShowFirst: 4,
				ShowLast:  4,
			},
		}),
		logmanager.WithTags("order", "transaction"),
		logmanager.WithExposeHeaders("Content-Type", "User-Agent"),
	)

	router := mux.NewRouter()
	router.Use(lmgorilla.Middleware(app))

	router.HandleFunc("/post/json", echoJSONHandler).Methods(http.MethodPost)

	fmt.Println("Gorilla Mux server running at http://localhost:8003")
	if err := http.ListenAndServe(":8003", router); err != nil {
		panic(err)
	}
}

func echoJSONHandler(w http.ResponseWriter, r *http.Request) {
	var body interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(body)
}