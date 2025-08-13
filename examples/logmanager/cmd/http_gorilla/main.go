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
		logmanager.WithMaskingConfig(
			[]logmanager.MaskingConfig{
				// Recursive wildcard masking - masks field at any level
				{
					Type:     logmanager.FullMask,
					JSONPath: "$..token",
				},
				{
					Type:     logmanager.FullMask,
					JSONPath: "$..password",
				},
				{
					Type:     logmanager.PartialMask,
					JSONPath: "$..apiKey",
					ShowFirst: 4,
					ShowLast:  4,
				},
			},
		),
		logmanager.WithTags("order", "transaction"),
		logmanager.WithExposeHeaders("Content-Type", "User-Agent"),
	)

	router := mux.NewRouter()
	router.Use(lmgorilla.Middleware(app))

	router.HandleFunc("/post/json", func(w http.ResponseWriter, r *http.Request) {
		// Read the body
		var body interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Write the body back to the response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(body)
	}).Methods(http.MethodPost)

	fmt.Println("Server is running at :8000")
	panic(http.ListenAndServe(":8000", router))
}
