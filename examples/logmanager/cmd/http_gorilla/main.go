package main

import (
	"examples/logmanager/internal/async"
	"examples/logmanager/internal/echo"
	"fmt"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgorilla"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	app := logmanager.NewApplication(
		logmanager.WithMaskConfigs(
			logmanager.MaskConfigs{
				{
					Field:     "credit_card",
					Type:      logmanager.PartialMask,
					ShowFirst: 4,
					ShowLast:  4,
				},
				{
					Field: "phone_number",
					Type:  logmanager.FullMask,
				},
				{
					Field: "email",
					Type:  logmanager.HideMask,
				},
			},
		),
		logmanager.WithTags("order", "transaction"),
		logmanager.WithExposeHeaders("Content-Type", "User-Agent"),
		//logmanager.WithDebug(),
		// logmanager.WithTraceIDKey("xid"), (optional) add xid if you want to change key of trace id
	)

	router := mux.NewRouter()
	router.Use(lmgorilla.Middleware(app))

	router.HandleFunc("/get/native", echo.Handler{Api: echo.NewAPINative("https://postman-echo.com")}.Get).Methods(http.MethodGet)
	router.HandleFunc("/get/resty", echo.Handler{Api: echo.NewApiResty("https://postman-echo.com")}.Get).Methods(http.MethodGet)
	router.HandleFunc("/get/async", async.Handler{}.Get).Methods(http.MethodGet)

	fmt.Println("Server is running at :8000")
	panic(http.ListenAndServe(":8000", router))
}
