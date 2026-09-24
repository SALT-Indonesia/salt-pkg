package lmgrpc_test

import (
	"context"
	"strings"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func newTraceTestApp(t *testing.T) (*logmanager.Application, *tracetest.SpanRecorder) {
	t.Helper()

	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	app := logmanager.NewApplication(
		logmanager.WithAppName("test-app"),
		logmanager.WithService("test-service"),
		logmanager.WithTracerProvider(provider),
	)

	return app, recorder
}

func sampleIDs() (trace.TraceID, trace.SpanID) {
	var traceID trace.TraceID
	for i := range traceID {
		traceID[i] = byte(0xa0 + i)
	}
	var spanID trace.SpanID
	for i := range spanID {
		spanID[i] = byte(0xb0 + i)
	}
	return traceID, spanID
}

// TestUnaryClientInterceptorInjectsTraceparent verifies that an outgoing gRPC
// call carries a W3C traceparent header for the span the interceptor created,
// so the callee can continue the same distributed trace.
func TestUnaryClientInterceptorInjectsTraceparent(t *testing.T) {
	app, recorder := newTraceTestApp(t)

	var captured metadata.MD
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		captured, _ = metadata.FromOutgoingContext(ctx)
		return nil
	}

	interceptor := lmgrpc.UnaryClientInterceptor(app)
	if err := interceptor(context.Background(), "/svc/Method", nil, nil, nil, invoker); err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	if captured == nil {
		t.Fatal("no outgoing metadata was set")
	}
	values := captured.Get("traceparent")
	if len(values) == 0 || values[0] == "" {
		t.Fatalf("outgoing metadata is missing traceparent: %v", captured)
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected exactly 1 recorded span, got %d", len(spans))
	}
	if want := spans[0].SpanContext().TraceID().String(); !strings.Contains(values[0], want) {
		t.Errorf("traceparent %q does not carry the transaction trace ID %s", values[0], want)
	}
}

// TestUnaryServerInterceptorJoinsIncomingTraceparent verifies that an inbound
// gRPC call carrying a traceparent header makes the server transaction join the
// caller's trace.
func TestUnaryServerInterceptorJoinsIncomingTraceparent(t *testing.T) {
	app, recorder := newTraceTestApp(t)

	traceID, spanID := sampleIDs()
	md := metadata.Pairs("traceparent", "00-"+traceID.String()+"-"+spanID.String()+"-01")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	var handlerCtx context.Context
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCtx = ctx
		return nil, nil
	}

	interceptor := lmgrpc.UnaryServerInterceptor(app)
	if _, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, handler); err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	if handlerCtx == nil {
		t.Fatal("handler was never invoked")
	}
	if got := trace.SpanContextFromContext(handlerCtx).TraceID(); got != traceID {
		t.Errorf("handler context trace ID = %s, want %s (the incoming trace)", got, traceID)
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected exactly 1 recorded span, got %d", len(spans))
	}
	root := spans[0]
	if got := root.SpanContext().TraceID(); got != traceID {
		t.Errorf("root span trace ID = %s, want %s (the incoming trace)", got, traceID)
	}
	if got := root.Parent().SpanID(); got != spanID {
		t.Errorf("root span parent span ID = %s, want %s (the incoming span)", got, spanID)
	}
}
