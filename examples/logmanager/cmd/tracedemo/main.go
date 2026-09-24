// tracedemo runs a small multi-service app in one process:
//
//	curl -> gateway (HTTP) -> svc-a (gRPC) -> svc-b (gRPC)
//
// Every service has its own logmanager Application. Run it with:
//
//	-otel=none      no OpenTelemetry at all (application trace_id only)
//	-otel=provider  WithTracerProvider -> OTLP/HTTP -> built-in stub collector
//	-otel=builtin   WithOpenTelemetry(...) -> OTLP/HTTP -> built-in stub collector
//
// No database or external collector is needed.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	coltrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/proto"
)

// ---------- observations recorded by each hop ----------

type hop struct {
	Service     string `json:"service"`
	CtxTraceID  any    `json:"ctx_trace_id"`
	TxTraceID   string `json:"tx_trace_id"`
	OTelTraceID string `json:"otel_trace_id"`
	HasDeadline bool   `json:"has_deadline"`
}

var (
	mu   sync.Mutex
	seen []hop
)

func observe(ctx context.Context, app *logmanager.Application, service string) {
	h := hop{Service: service, CtxTraceID: ctx.Value(app.TraceIDContextKey())}
	if tx := logmanager.FromContext(ctx); tx != nil {
		h.TxTraceID = tx.TraceID()
	}
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		h.OTelTraceID = sc.TraceID().String()
	}
	_, h.HasDeadline = ctx.Deadline()
	mu.Lock()
	seen = append(seen, h)
	mu.Unlock()
}

// ---------- stub OTLP/HTTP collector ----------

type collected struct {
	Service string `json:"service"`
	Name    string `json:"name"`
	TraceID string `json:"trace_id"`
	SpanID  string `json:"span_id"`
	Parent  string `json:"parent_span_id"`
	AppID   string `json:"logmanager.trace_id"`
}

var (
	cmu   sync.Mutex
	spans []collected
)

func collector(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	var req coltrace.ExportTraceServiceRequest
	if err := proto.Unmarshal(b, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cmu.Lock()
	defer cmu.Unlock()
	for _, rs := range req.ResourceSpans {
		svc := ""
		for _, kv := range rs.Resource.GetAttributes() {
			if kv.Key == "service.name" {
				svc = kv.Value.GetStringValue()
			}
		}
		for _, ss := range rs.ScopeSpans {
			for _, sp := range ss.Spans {
				c := collected{Service: svc, Name: sp.Name, TraceID: fmt.Sprintf("%x", sp.TraceId), SpanID: fmt.Sprintf("%x", sp.SpanId), Parent: fmt.Sprintf("%x", sp.ParentSpanId)}
				for _, kv := range sp.Attributes {
					if kv.Key == "logmanager.trace_id" {
						c.AppID = kv.Value.GetStringValue()
					}
				}
				spans = append(spans, c)
			}
		}
	}
}

// ---------- services ----------

type health struct {
	grpc_health_v1.UnimplementedHealthServer
	fn func(ctx context.Context) error
}

func (h health) Check(ctx context.Context, _ *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, h.fn(ctx)
}

func serve(app *logmanager.Application, fn func(ctx context.Context) error) string {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(lmgrpc.UnaryServerInterceptor(app)))
	grpc_health_v1.RegisterHealthServer(s, health{fn: fn})
	go func() { _ = s.Serve(lis) }()
	return lis.Addr().String()
}

func dial(app *logmanager.Application, addr string) grpc_health_v1.HealthClient {
	cc, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(lmgrpc.UnaryClientInterceptor(app)))
	if err != nil {
		log.Fatal(err)
	}
	return grpc_health_v1.NewHealthClient(cc)
}

func main() {
	mode := flag.String("otel", "none", "none | provider | builtin")
	addr := flag.String("addr", "127.0.0.1:18080", "gateway listen address")
	flag.Parse()

	collectorSrv := &http.Server{Handler: http.HandlerFunc(collector)}
	cl, _ := net.Listen("tcp", "127.0.0.1:0")
	go func() { _ = collectorSrv.Serve(cl) }()
	collectorURL := "http://" + cl.Addr().String()

	newApp := func(name string) *logmanager.Application {
		opts := []logmanager.Option{logmanager.WithAppName(name), logmanager.WithService(name)}
		switch *mode {
		case "provider":
			exp, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpointURL(collectorURL+"/v1/traces"))
			if err != nil {
				log.Fatal(err)
			}
			tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp, sdktrace.WithBatchTimeout(100*time.Millisecond)),
				sdktrace.WithResource(resource.NewSchemaless(attribute.String("service.name", name))))
			opts = append(opts, logmanager.WithTracerProvider(tp))
		case "builtin":
			opts = append(opts, logmanager.WithOpenTelemetry(
				logmanager.WithOTelEndpoint(cl.Addr().String()), logmanager.WithOTelProtocol("http/protobuf"), logmanager.WithOTelInsecure()))
		}
		return logmanager.NewApplication(opts...)
	}
	gw, appA, appB := newApp("gateway"), newApp("svc-a"), newApp("svc-b")

	addrB := serve(appB, func(ctx context.Context) error { observe(ctx, appB, "svc-b"); return nil })
	clientB := dial(appA, addrB)
	addrA := serve(appA, func(ctx context.Context) error {
		observe(ctx, appA, "svc-a")
		_, err := clientB.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		return err
	})
	clientA := dial(gw, addrA)

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		// Join the caller's trace when a traceparent header is present.
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		tx := gw.StartHttpWithContext(ctx, r.Header.Get(gw.TraceIDHeaderKey()), r.Method+" "+r.URL.Path)
		defer tx.End()

		ctx = context.WithValue(ctx, gw.TraceIDContextKey(), tx.TraceID())
		ctx, cancel := context.WithTimeout(tx.ToContext(ctx), 5*time.Second)
		defer cancel()
		observe(ctx, gw, "gateway")

		if _, err := clientA.Check(ctx, &grpc_health_v1.HealthCheckRequest{}); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		_, _ = fmt.Fprintln(w, "ok")
	})
	// /report returns (and clears) the hops observed by the services plus the
	// spans received by the stub collector.
	mux.HandleFunc("/report", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		cmu.Lock()
		sort.SliceStable(spans, func(i, j int) bool { return spans[i].Service < spans[j].Service })
		out := map[string]any{"mode": *mode, "hops": seen, "collector_spans": spans}
		seen, spans = nil, nil
		cmu.Unlock()
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
	})

	log.Printf("tracedemo mode=%s listening on http://%s (collector %s)", *mode, *addr, collectorURL)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
