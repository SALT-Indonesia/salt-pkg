package logmanager_test

import (
	"context"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

type ctxKey struct{}

// ToContext must keep the values, deadline and cancellation of the context it is
// given even when the transaction owns an OpenTelemetry span.
func TestToContextWithOTelPreservesCallerContext(t *testing.T) {
	app, _ := newRecorderApp(t)

	parent, cancel := context.WithCancel(context.WithValue(context.Background(), ctxKey{}, "value"))
	tx := app.StartConsumer("11111111-2222-3333-4444-555555555555")
	defer tx.End()

	ctx := tx.ToContext(parent)

	assert.Equal(t, "value", ctx.Value(ctxKey{}))
	assert.Same(t, tx, logmanager.FromContext(ctx))

	cancel()
	assert.ErrorIs(t, ctx.Err(), context.Canceled)
}

func TestToContextWithOTelEmbedsSpanContext(t *testing.T) {
	app, _ := newRecorderApp(t)

	tx := app.StartHttp("11111111-2222-3333-4444-555555555555", "GET /ping")
	defer tx.End()

	ctx := tx.ToContext(context.Background())

	assert.True(t, trace.SpanContextFromContext(ctx).IsValid())
}

func TestToContextWithoutOTelPreservesCallerContext(t *testing.T) {
	app := logmanager.NewApplication(logmanager.WithAppName("test-app"))

	parent, cancel := context.WithCancel(context.WithValue(context.Background(), ctxKey{}, "value"))
	tx := app.StartConsumer("11111111-2222-3333-4444-555555555555")
	defer tx.End()

	ctx := tx.ToContext(parent)

	assert.Equal(t, "value", ctx.Value(ctxKey{}))
	cancel()
	assert.ErrorIs(t, ctx.Err(), context.Canceled)
}
