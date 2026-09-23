package lmgrpc

import (
	"context"
	"encoding/json"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// metadataCarrier adapts gRPC metadata to the OpenTelemetry TextMapCarrier
// interface so a W3C trace context can be injected into, and extracted from,
// gRPC metadata.
type metadataCarrier struct {
	md metadata.MD
}

func (c metadataCarrier) Get(key string) string {
	values := c.md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c metadataCarrier) Set(key, value string) {
	c.md.Set(key, value)
}

func (c metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(c.md))
	for k := range c.md {
		keys = append(keys, k)
	}
	return keys
}

// extractTraceContext pulls the incoming W3C trace context (traceparent) out of
// the gRPC metadata and returns a context carrying it. The context is returned
// unchanged when no trace context is present, so this is always safe to call.
func extractTraceContext(ctx context.Context, md metadata.MD) context.Context {
	if md == nil {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, metadataCarrier{md: md})
}

// injectTraceContext writes the trace context found in ctx into md so the
// callee can continue the same distributed trace.
func injectTraceContext(ctx context.Context, md metadata.MD) {
	if md == nil {
		return
	}
	otel.GetTextMapPropagator().Inject(ctx, metadataCarrier{md: md})
}

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

		// Continue the caller's distributed trace when one is present. The
		// extracted context is handed to StartHttpWithContext so the
		// transaction's root span becomes a child of the caller's span instead
		// of starting a brand new trace.
		ctx = extractTraceContext(ctx, md)

		tx := app.StartHttpWithContext(ctx, traceID, info.FullMethod)
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

		// The transaction must exist before the trace context is injected,
		// because the injected span is the transaction's own span.
		tx := app.StartHttpWithContext(ctx, traceID, method)
		tx.SetRequestValue(req)

		// set trace id to context
		ctx = context.WithValue(ctx, app.TraceIDContextKey(), traceID)
		ctx = tx.ToContext(ctx)

		// Create outgoing metadata with the trace ID and propagate the W3C
		// trace context so the callee joins the same distributed trace.
		md := metadata.New(map[string]string{
			app.TraceIDHeaderKey(): traceID,
		})
		injectTraceContext(ctx, md)
		ctx = metadata.NewOutgoingContext(ctx, md)

		// Invoke the RPC
		err := invoker(ctx, method, req, reply, cc, opts...)
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

		tx := app.StartHttpWithContext(ctx, traceID, method)

		// set trace id to context
		ctx = context.WithValue(ctx, app.TraceIDContextKey(), traceID)
		ctx = tx.ToContext(ctx)

		// Create outgoing metadata with the trace ID and propagate the W3C
		// trace context so the callee joins the same distributed trace.
		md := metadata.New(map[string]string{
			app.TraceIDHeaderKey(): traceID,
		})
		injectTraceContext(ctx, md)
		ctx = metadata.NewOutgoingContext(ctx, md)

		// Create the stream
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
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

		// Continue the caller's distributed trace when one is present.
		ctx = extractTraceContext(ctx, md)

		tx := app.StartHttpWithContext(ctx, traceID, info.FullMethod)

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
