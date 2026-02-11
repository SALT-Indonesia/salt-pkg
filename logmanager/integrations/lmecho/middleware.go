package lmecho

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"net/http"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/labstack/echo/v4"
)

type customWriter struct {
	http.ResponseWriter
	Body *bytes.Buffer
}

func (w *customWriter) Write(b []byte) (int, error) {
	w.Body.Write(b)

	return w.ResponseWriter.Write(b)
}

// Flush implements http.Flusher interface for streaming responses (SSE, chunked)
func (w *customWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements http.Hijacker interface for WebSocket upgrades
func (w *customWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.New("http.Hijacker interface not supported")
}

// Push implements http.Pusher interface for HTTP/2 server push
func (w *customWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

func traceID(c echo.Context, app *logmanager.Application) string {
	if c.Request() == nil {
		return ""
	}

	traceID := ""
	contextValue := c.Request().Context().Value(app.TraceIDContextKey())
	if contextValue != nil {
		traceID = contextValue.(string)
	}
	if app.TraceIDViaHeader() {
		traceID = c.Request().Header.Get(app.TraceIDHeaderKey())
	}

	return traceID
}

func routeName(r *http.Request) string {
	if r == nil {
		return "NotFoundHandler"
	}
	return r.Method + " " + r.URL.String()
}

func writeResponse(next echo.HandlerFunc, c echo.Context, tx *logmanager.Transaction, traceIDKey, traceID string) error {
	cw := &customWriter{c.Response().Writer, new(bytes.Buffer)}
	cw.Header().Set(traceIDKey, traceID)
	c.Response().Writer = cw
	err := next(c)

	tx.SetWebResponse(logmanager.WebResponse{
		StatusCode: c.Response().Status,
		Body:       cw.Body.Bytes(),
	})

	return err
}

func updateRequest(c echo.Context, app *logmanager.Application, traceID string, tx *logmanager.Transaction) {
	r := c.Request()
	r = logmanager.RequestWithContext(r, app.TraceIDContextKey(), traceID)
	r = logmanager.RequestWithTransactionContext(r, tx)

	c.SetRequest(r)
}

func Middleware(app *logmanager.Application) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			traceId := traceID(c, app)
			tx := app.StartHttp(traceId, routeName(c.Request()))
			defer tx.End()

			tx.SetWebRequest(c.Request())

			// Use the actual trace ID from the transaction (maybe auto-generated)
			actualTraceID := tx.TraceID()
			updateRequest(c, app, actualTraceID, tx)

			err := writeResponse(next, c, tx, app.TraceIDHeaderKey(), actualTraceID)

			return err
		}
	}
}
