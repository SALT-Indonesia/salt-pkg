package lmgrpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgrpc"
	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type healthFunc func(ctx context.Context) error

type healthServer struct {
	grpc_health_v1.UnimplementedHealthServer
	check healthFunc
}

func (h healthServer) Check(ctx context.Context, _ *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{}, h.check(ctx)
}

func startHealthServer(t *testing.T, app *logmanager.Application, check healthFunc) string {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)

	srv := grpc.NewServer(grpc.UnaryInterceptor(lmgrpc.UnaryServerInterceptor(app)))
	grpc_health_v1.RegisterHealthServer(srv, healthServer{check: check})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	return lis.Addr().String()
}

func newHealthClient(t *testing.T, app *logmanager.Application, addr string) grpc_health_v1.HealthClient {
	t.Helper()

	cc, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(lmgrpc.UnaryClientInterceptor(app)),
	)
	assert.NoError(t, err)
	t.Cleanup(func() { _ = cc.Close() })

	return grpc_health_v1.NewHealthClient(cc)
}

// runTraceIDChain drives caller -> service A -> service B over real gRPC and
// returns the application trace IDs observed inside each handler.
func runTraceIDChain(t *testing.T, newApp func(name string) *logmanager.Application, want string) (aCtx, aTx, bCtx, bTx any) {
	t.Helper()

	caller, svcA, svcB := newApp("caller"), newApp("svc-a"), newApp("svc-b")

	addrB := startHealthServer(t, svcB, func(ctx context.Context) error {
		bCtx = ctx.Value(svcB.TraceIDContextKey())
		if tx := logmanager.FromContext(ctx); tx != nil {
			bTx = tx.TraceID()
		}
		return nil
	})
	clientB := newHealthClient(t, svcA, addrB)

	addrA := startHealthServer(t, svcA, func(ctx context.Context) error {
		aCtx = ctx.Value(svcA.TraceIDContextKey())
		if tx := logmanager.FromContext(ctx); tx != nil {
			aTx = tx.TraceID()
		}
		_, err := clientB.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		return err
	})
	clientA := newHealthClient(t, caller, addrA)

	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), caller.TraceIDContextKey(), want), 5*time.Second)
	defer cancel()

	_, err := clientA.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	assert.NoError(t, err)

	return aCtx, aTx, bCtx, bTx
}

// The application-level trace ID must travel caller -> A -> B when OpenTelemetry
// is not used at all.
func TestTraceIDPropagatesAcrossServicesWithoutOTel(t *testing.T) {
	const want = "11111111-2222-3333-4444-555555555555"

	aCtx, aTx, bCtx, bTx := runTraceIDChain(t, func(name string) *logmanager.Application {
		return logmanager.NewApplication(logmanager.WithAppName(name), logmanager.WithService(name))
	}, want)

	assert.Equal(t, want, aCtx)
	assert.Equal(t, want, aTx)
	assert.Equal(t, want, bCtx)
	assert.Equal(t, want, bTx)
}

// With OpenTelemetry enabled both IDs must propagate: the application-level
// trace ID and the OpenTelemetry trace ID (one shared trace across every hop).
func TestTraceIDPropagatesAcrossServicesWithOTel(t *testing.T) {
	const want = "11111111-2222-3333-4444-555555555555"

	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	aCtx, aTx, bCtx, bTx := runTraceIDChain(t, func(name string) *logmanager.Application {
		return logmanager.NewApplication(
			logmanager.WithAppName(name),
			logmanager.WithService(name),
			logmanager.WithTracerProvider(provider),
		)
	}, want)

	assert.Equal(t, want, aCtx)
	assert.Equal(t, want, aTx)
	assert.Equal(t, want, bCtx)
	assert.Equal(t, want, bTx)

	// caller client, A server, A client, B server: all in a single trace.
	assert.Eventually(t, func() bool { return len(recorder.Ended()) == 4 }, 2*time.Second, 10*time.Millisecond)
	traceIDs := map[string]struct{}{}
	for _, span := range recorder.Ended() {
		traceIDs[span.SpanContext().TraceID().String()] = struct{}{}
	}
	assert.Len(t, traceIDs, 1)
}
