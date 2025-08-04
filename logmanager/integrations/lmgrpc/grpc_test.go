package lmgrpc_test

import (
	"context"
	"errors"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgrpc"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestUnaryServerInterceptorSuite(t *testing.T) {
	suite.Run(t, new(UnaryServerInterceptorSuite))
}

// UnaryServerInterceptorSuite defines a test suite for UnaryServerInterceptor
type UnaryServerInterceptorSuite struct {
	suite.Suite
	app         *logmanager.TestableApplication
	interceptor grpc.UnaryServerInterceptor
}

// SetupTest initializes the app and interceptor before each test
func (suite *UnaryServerInterceptorSuite) SetupTest() {
	suite.app = logmanager.NewTestableApplication()

	suite.interceptor = lmgrpc.UnaryServerInterceptor(suite.app.Application)
}

func (suite *UnaryServerInterceptorSuite) TestWithTraceIDInMetadata() {
	// Reset logged entries before test
	suite.app.ResetLoggedEntries()

	md := metadata.Pairs(suite.app.TraceIDHeaderKey(), "mock-trace-id")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	req := "mock-request"

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		assert.Equal(suite.T(), "mock-trace-id", ctx.Value(suite.app.TraceIDContextKey()))
		return "mock-response", nil
	}

	resp, err := suite.interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/mock/Method"}, handler)

	suite.NoError(err)
	suite.Equal("mock-response", resp)

	// Assert logged data keys and values
	suite.Equal(1, suite.app.CountLoggedEntries(), "Should have exactly one logged entry")

	// Verify essential logged fields exist
	suite.True(suite.app.HasLoggedField("trace_id"), "Should log trace_id field")
	suite.True(suite.app.HasLoggedField("name"), "Should log name field")
	suite.True(suite.app.HasLoggedField("type"), "Should log type field")
	suite.True(suite.app.HasLoggedField("start"), "Should log start field")
	suite.True(suite.app.HasLoggedField("latency"), "Should log latency field")
	suite.True(suite.app.HasLoggedField("service"), "Should log service field")
	suite.True(suite.app.HasLoggedField("status"), "Should log status field")

	// Verify logged field values
	suite.Equal("mock-trace-id", suite.app.GetLoggedField("trace_id"), "Should log correct trace_id")
	suite.Equal("/mock/Method", suite.app.GetLoggedField("name"), "Should log correct method name")
	suite.Equal(logmanager.TxnTypeHttp, suite.app.GetLoggedField("type"), "Should log HTTP transaction type")
	suite.Equal("default", suite.app.GetLoggedField("service"), "Should log default service name")
	// Note: gRPC interceptor doesn't log separate method field, it's included in name
	suite.Equal(200, suite.app.GetLoggedField("status"), "Should log success status code")

	// Verify log level is Info for successful requests
	suite.Equal(logrus.InfoLevel, suite.app.GetLoggedLevel(), "Should log at Info level for successful requests")
	suite.Equal("", suite.app.GetLoggedMessage(), "Should have empty message for Info level logs")
}

func (suite *UnaryServerInterceptorSuite) TestFallbackToUUID() {
	ctx := context.WithValue(
		context.Background(),
		suite.app.TraceIDContextKey(),
		"mock-trace-id",
	)
	req := "mock-request"

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		traceID := ctx.Value(suite.app.TraceIDContextKey())
		suite.NotNil(traceID)
		return "mock-response", nil
	}

	resp, err := suite.interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/mock/Method"}, handler)

	suite.NoError(err)
	suite.Equal("mock-response", resp)
}

func (suite *UnaryServerInterceptorSuite) TestErrorInHandler() {
	ctx := context.Background()
	req := "mock-request"
	mockError := errors.New("mock error")

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, mockError
	}

	resp, err := suite.interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/mock/Method"}, handler)

	suite.Nil(resp)
	suite.Equal(mockError, err)
}

func (suite *UnaryServerInterceptorSuite) TestResponseSerializationError() {
	ctx := context.Background()
	req := "mock-request"

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Response that will fail JSON serialization
		return func() {}, nil
	}

	_, err := suite.interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/mock/Method"}, handler)

	suite.NoError(err)
}

func (suite *UnaryServerInterceptorSuite) TestMetadataExtractedSuccessfully() {
	md := metadata.Pairs(string(suite.app.TraceIDContextKey()), "mock-trace-id")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	req := "mock-request"

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		traceID := ctx.Value(suite.app.TraceIDContextKey())
		suite.Equal("mock-trace-id", traceID)
		return nil, nil
	}

	_, err := suite.interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/mock/Method"}, handler)
	suite.NoError(err)
}
