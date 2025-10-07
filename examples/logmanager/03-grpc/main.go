package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "examples/logmanager/03-grpc/proto/proto"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgrpc"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GreeterServer struct {
	pb.GreeterServer
}

func (s *GreeterServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	// Simulate processing time
	time.Sleep(230 * time.Millisecond)

	// Simulate error for demonstration
	if req.GetName() == "error" {
		return nil, status.New(codes.InvalidArgument, "invalid name provided").Err()
	}

	return &pb.HelloReply{
		Message: fmt.Sprintf("Hello %s!", req.GetName()),
	}, nil
}

func main() {
	viper.SetDefault("port", "50051")
	port := viper.GetString("port")

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	app := logmanager.NewApplication(
		logmanager.WithAppName("grpc-server"),
		logmanager.WithTraceIDHeaderKey("X-Trace-Id"),
	)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			lmgrpc.UnaryServerInterceptor(app),
		),
		grpc.StreamInterceptor(
			lmgrpc.StreamServerInterceptor(app),
		),
	)

	pb.RegisterGreeterServer(grpcServer, &GreeterServer{})

	fmt.Printf("gRPC server running at http://localhost:%s\n", port)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}