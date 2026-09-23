package logmanager_test

import (
	"context"
	"strings"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// incomingTraceContext reproduces what a caller's traceparent header looks like
// once it has been turned into a context by a propagator Extract.
func incomingTraceContext() (context.Context, trace.SpanContext) {
	var traceID trace.TraceID
	for i := range traceID {
		traceID[i] = byte(0xa0 + i)
	}
	var spanID trace.SpanID
	for i := range spanID {
		spanID[i] = byte(0xb0 + i)
	}

	ctx := propagation.TraceContext{}.Extract(context.Background(), propagation.MapCarrier{
		"traceparent": "00-" + traceID.String() + "-" + spanID.String() + "-01",
	})

	return ctx, trace.SpanContextFromContext(ctx)
}

func newRecorderApp(t *testing.T) (*logmanager.Application, *tracetest.SpanRecorder) {
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

// TestStartHttpWithContextJoinsIncomingTrace is the regression test for the
// root span hardcoding context.Background(): when a context carrying an inbound
// trace context is passed, the transaction must join that trace instead of
// starting a brand new one.
func TestStartHttpWithContextJoinsIncomingTrace(t *testing.T) {
	app, recorder := newRecorderApp(t)
	ctx, incoming := incomingTraceContext()

	if !incoming.IsValid() {
		t.Fatal("test setup: incoming span context is not valid")
	}

	const appTraceID = "11111111-2222-3333-4444-555555555555"
	tx := app.StartHttpWithContext(ctx, appTraceID, "GET /ping")
	tx.End()

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected exactly 1 recorded span, got %d", len(spans))
	}

	root := spans[0]
	if got := root.SpanContext().TraceID(); got != incoming.TraceID() {
		t.Errorf("root span trace ID = %s, want %s (the incoming trace)", got, incoming.TraceID())
	}
	if got := root.Parent().SpanID(); got != incoming.SpanID() {
		t.Errorf("root span parent span ID = %s, want %s (the incoming span)", got, incoming.SpanID())
	}

	// The application-level trace ID must be exported as a span attribute,
	// otherwise the log line and the span cannot be correlated.
	var found bool
	for _, kv := range root.Attributes() {
		if string(kv.Key) == "logmanager.trace_id" && kv.Value.AsString() == appTraceID {
			found = true
		}
	}
	if !found {
		t.Errorf("root span is missing logmanager.trace_id=%s, attributes: %v", appTraceID, root.Attributes())
	}
}

// TestStartHttpWithoutContextStartsNewTrace locks in the backward-compatible
// behaviour of the legacy API.
func TestStartHttpWithoutContextStartsNewTrace(t *testing.T) {
	app, recorder := newRecorderApp(t)

	tx := app.StartHttp("11111111-2222-3333-4444-555555555555", "GET /ping")
	tx.End()

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected exactly 1 recorded span, got %d", len(spans))
	}
	if spans[0].Parent().IsValid() {
		t.Errorf("StartHttp must start a new root span, got parent %s", spans[0].Parent().SpanID())
	}
}

// TestTransactionToContextCarriesSpan ensures the OTel span is embedded in the
// transaction context so downstream code can Inject it into an outgoing carrier.
func TestTransactionToContextCarriesSpan(t *testing.T) {
	app, _ := newRecorderApp(t)
	ctx, incoming := incomingTraceContext()

	tx := app.StartHttpWithContext(ctx, "11111111-2222-3333-4444-555555555555", "GET /ping")
	defer tx.End()

	carrier := propagation.MapCarrier{}
	propagation.TraceContext{}.Inject(tx.ToContext(context.Background()), carrier)

	traceparent, ok := carrier["traceparent"]
	if !ok || traceparent == "" {
		t.Fatalf("traceparent was not injected from the transaction context: %v", carrier)
	}
	if want := incoming.TraceID().String(); !strings.Contains(traceparent, want) {
		t.Errorf("traceparent %q does not carry trace ID %s", traceparent, want)
	}
	if logmanager.FromContext(tx.ToContext(context.Background())) != tx {
		t.Error("transaction is no longer retrievable from its own context")
	}
}
