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

func UnaryClientInterceptor(app *logmanager.Application) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Extract trace ID from context
		var traceID string
		if val := ctx.Value(app.TraceIDContextKey()); val != nil {
			traceID = val.(string)
		}

		if traceID == "" {
			traceID = uuid.NewString()
		}

		// Create outgoing metadata with trace ID
		md := metadata.New(map[string]string{
			app.TraceIDHeaderKey(): traceID,
		})
		ctx = metadata.NewOutgoingContext(ctx, md)

		tx := app.StartHttp(traceID, method)
		tx.SetRequestValue(req)

		// set trace id to context
		ctx = context.WithValue(ctx, app.TraceIDContextKey(), traceID)

		// Invoke the RPC
		err := invoker(tx.ToContext(ctx), method, req, reply, cc, opts...)
		defer tx.End()

		// Log the response
		var respJSON []byte
		if reply != nil {
			respJSON, _ = json.Marshal(reply)
		}

		tx.SetWebResponse(logmanager.WebResponse{
			StatusCode: ConvertCodeToHTTPStatus(status.Code(err)),
			Body:       respJSON,
		})

		return err
	}
}

func StreamClientInterceptor(app *logmanager.Application) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// Extract trace ID from context
		var traceID string
		if val := ctx.Value(app.TraceIDContextKey()); val != nil {
			traceID = val.(string)
		}

		if traceID == "" {
			traceID = uuid.NewString()
		}

		// Create outgoing metadata with trace ID
		md := metadata.New(map[string]string{
			app.TraceIDHeaderKey(): traceID,
		})
		ctx = metadata.NewOutgoingContext(ctx, md)

		tx := app.StartHttp(traceID, method)

		// set trace id to context
		ctx = context.WithValue(ctx, app.TraceIDContextKey(), traceID)

		// Create the stream
		clientStream, err := streamer(tx.ToContext(ctx), desc, cc, method, opts...)
		if err != nil {
			tx.SetWebResponse(logmanager.WebResponse{
				StatusCode: ConvertCodeToHTTPStatus(status.Code(err)),
			})
			tx.End()
			return nil, err
		}

		// Wrap the client stream to track completion
		wrappedStream := &wrappedClientStream{
			ClientStream: clientStream,
			tx:           tx,
		}

		return wrappedStream, nil
	}
}

type wrappedClientStream struct {
	grpc.ClientStream
	tx *logmanager.Transaction
}

func (w *wrappedClientStream) SendMsg(m interface{}) error {
	err := w.ClientStream.SendMsg(m)
	if err != nil {
		w.tx.SetWebResponse(logmanager.WebResponse{
			StatusCode: ConvertCodeToHTTPStatus(status.Code(err)),
		})
	}
	return err
}

func (w *wrappedClientStream) RecvMsg(m interface{}) error {
	err := w.ClientStream.RecvMsg(m)
	if err != nil {
		w.tx.SetWebResponse(logmanager.WebResponse{
			StatusCode: ConvertCodeToHTTPStatus(status.Code(err)),
		})
		w.tx.End()
	}
	return err
}

func StreamServerInterceptor(app *logmanager.Application) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Extract metadata (headers) from incoming context
		ctx := ss.Context()
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

		// set trace id to context
		ctx = context.WithValue(ctx, app.TraceIDContextKey(), traceID)

		// Wrap the server stream to track messages
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          tx.ToContext(ctx),
			tx:           tx,
		}

		// Handle the stream
		err := handler(srv, wrappedStream)

		tx.SetWebResponse(logmanager.WebResponse{
			StatusCode: ConvertCodeToHTTPStatus(status.Code(err)),
		})
		tx.End()

		return err
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
	tx  *logmanager.Transaction
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func (w *wrappedServerStream) SendMsg(m interface{}) error {
	err := w.ServerStream.SendMsg(m)
	if err != nil {
		w.tx.SetWebResponse(logmanager.WebResponse{
			StatusCode: ConvertCodeToHTTPStatus(status.Code(err)),
		})
	}
	return err
}

func (w *wrappedServerStream) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)
	if err != nil {
		w.tx.SetWebResponse(logmanager.WebResponse{
			StatusCode: ConvertCodeToHTTPStatus(status.Code(err)),
		})
	}
	return err
}
