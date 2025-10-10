package lmgin

import (
	"bytes"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func routeName(c *gin.Context) string {
	if nil == c.Request {
		return "NotFoundHandler"
	}

	return c.Request.Method + " " + c.Request.URL.String()
}

func Middleware(app *logmanager.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		var traceID string
		if app.TraceIDViaHeader() {
			traceID = c.Request.Header.Get(app.TraceIDHeaderKey())
		} else {
			traceID = c.GetString(string(app.TraceIDContextKey()))
		}

		if traceID == "" {
			traceID = uuid.NewString()
		}

		c.Set(string(app.TraceIDContextKey()), traceID)

		tx := app.StartHttp(traceID, routeName(c))
		tx.SetWebRequest(c.Request)

		c.Set(logmanager.TransactionContextKey.String(), tx)

		// Also propagate transaction to c.Request.Context() for downstream layers
		c.Request = logmanager.RequestWithTransactionContext(c.Request, tx)

		rw := &responseCapture{Body: new(bytes.Buffer), ResponseWriter: c.Writer}
		rw.Header().Set(app.TraceIDHeaderKey(), traceID)
		c.Writer = rw

		c.Next()

		tx.SetWebResponse(logmanager.WebResponse{
			StatusCode: rw.StatusCode,
			Body:       rw.Body.Bytes(),
		})
		tx.End()
	}
}

type responseCapture struct {
	gin.ResponseWriter
	Body       *bytes.Buffer
	StatusCode int
}

func (r *responseCapture) Write(b []byte) (int, error) {
	r.Body.Write(b)                  // Capture the data in our buffer
	return r.ResponseWriter.Write(b) // Write the data to the actual response
}

func (r *responseCapture) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
