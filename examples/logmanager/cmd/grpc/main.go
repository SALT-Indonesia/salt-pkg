package main

import (
	"context"
	pb "examples/logmanager/proto/proto"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgrpc"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"time"
)

type server struct {
	pb.GreeterServer
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	time.Sleep(230 * time.Millisecond)

	return nil, status.New(codes.InvalidArgument, "this is error").Err()

	//return nil, errors.New("this is error")

	//return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	// Set default port value
	viper.SetDefault("port", "50051")

	// Retrieve the port from configuration
	port := viper.GetString("port")

	// Define the address for the gRPC server using the configured port
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port 50051: %v", err)
	}

	// start implement log manager here
	app := logmanager.NewApplication(
		logmanager.WithAppName("grpc"),
		logmanager.WithTraceIDHeaderKey("X-Trace-Id"),
	)

	// Create a new gRPC server instance.
	grpcServer := grpc.NewServer(
		// inject middleware log manager
		grpc.UnaryInterceptor(
			lmgrpc.UnaryServerInterceptor(app),
		),
	)

	// Register your services here.
	// Example:
	pb.RegisterGreeterServer(grpcServer, &server{})

	log.Println("gRPC server is running on port http://localhost:50051")

	// Start serving incoming connections.
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
