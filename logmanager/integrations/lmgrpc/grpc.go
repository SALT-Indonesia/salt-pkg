package lmgrpc

import (
	"context"
	"encoding/json"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor(app *logmanager.Application) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract metadata (headers) from incoming context
		md, _ := metadata.FromIncomingContext(ctx)
		var traceID string
		traceIDs := md.Get(app.TraceIDHeaderKey())
		if len(traceIDs) > 0 {
			traceID = traceIDs[0]
		}

		if traceID == "" {
			traceIDs = md.Get(string(app.TraceIDContextKey()))
			if len(traceIDs) > 0 {
				traceID = traceIDs[0]
			}
		}

		if traceID == "" {
			traceID = uuid.NewString()
		}

		tx := app.StartHttp(traceID, info.FullMethod)
		tx.SetRequestValue(req)

		// set trace id to context
		ctx = context.WithValue(ctx, app.TraceIDContextKey(), traceID)

		// Process the request by invoking the actual handler
		resp, err := handler(tx.ToContext(ctx), req)
		defer tx.End()

		// Log the response
		var respJSON []byte
		if resp != nil {
			respJSON, _ = json.Marshal(resp)
		}

		tx.SetWebResponse(logmanager.WebResponse{
			StatusCode: ConvertCodeToHTTPStatus(status.Code(err)),
			Body:       respJSON,
		})

		return resp, err
	}
}
