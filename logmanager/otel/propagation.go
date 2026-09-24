package otel

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// EnsureW3CPropagator installs the W3C Trace Context propagator (together with
// Baggage) as the global OpenTelemetry propagator, but only when no propagator
// has been configured yet.
//
// The OpenTelemetry default global propagator is a no-op, so without this
// Inject() writes nothing into an outgoing carrier and Extract() never restores
// a trace context: distributed traces silently break at every process
// boundary. logmanager calls this as soon as tracing is enabled, so the
// traceparent header flows between services out of the box.
//
// Applications that prefer another propagator (B3, Jaeger, ...) can call
// otel.SetTextMapPropagator themselves: this function leaves an already
// configured propagator untouched.
func EnsureW3CPropagator() {
	if len(otel.GetTextMapPropagator().Fields()) > 0 {
		// A propagator with fields is already configured, so respect it.
		return
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
}
