package lmgorilla

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func routeName(r *http.Request) string {
	route := mux.CurrentRoute(r)
	if nil == route {
		return "NotFoundHandler"
	}
	if n := route.GetName(); n != "" {
		return n
	}
	if n, _ := route.GetPathTemplate(); n != "" {
		return r.Method + " " + n
	}
	n, _ := route.GetHostTemplate()
	return r.Method + " " + n
}

func Middleware(app *logmanager.Application) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var traceID string
			if app.TraceIDViaHeader() {
				traceID = r.Header.Get(app.TraceIDHeaderKey())
			} else {
				traceID, _ = r.Context().Value(app.TraceIDContextKey()).(string)
			}

			if traceID == "" {
				traceID = uuid.NewString()
			}

			tx := app.StartHttp(traceID, routeName(r))
			defer tx.End()

			tx.SetWebRequest(r)
			w = tx.SetWebResponseHttp(w)
			w.Header().Set(app.TraceIDHeaderKey(), traceID)
			r = logmanager.RequestWithContext(r, app.TraceIDContextKey(), traceID)
			r = logmanager.RequestWithTransactionContext(r, tx)
			next.ServeHTTP(w, r)
		})
	}
}
