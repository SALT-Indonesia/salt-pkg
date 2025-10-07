package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "examples/logmanager/03-grpc/proto/proto"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgrpc"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func runClient() {
	viper.SetDefault("server_address", "localhost:50051")
	serverAddress := viper.GetString("server_address")

	// Initialize logmanager for client
	app := logmanager.NewApplication(
		logmanager.WithAppName("grpc-client"),
		logmanager.WithTraceIDHeaderKey("X-Trace-Id"),
	)

	// Create gRPC connection with client interceptors
	conn, err := grpc.NewClient(
		serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(lmgrpc.UnaryClientInterceptor(app)),
		grpc.WithStreamInterceptor(lmgrpc.StreamClientInterceptor(app)),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)

	// Example 1: Successful request
	fmt.Println("=== Example 1: Successful Request ===")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "World"})
	if err != nil {
		log.Printf("Error calling SayHello: %v", err)
	} else {
		fmt.Printf("Response: %s\n", resp.GetMessage())
	}

	time.Sleep(100 * time.Millisecond)

	// Example 2: Error request
	fmt.Println("\n=== Example 2: Error Request ===")
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()

	resp2, err2 := client.SayHello(ctx2, &pb.HelloRequest{Name: "error"})
	if err2 != nil {
		log.Printf("Expected error received: %v", err2)
	} else {
		fmt.Printf("Response: %s\n", resp2.GetMessage())
	}

	time.Sleep(100 * time.Millisecond)

	// Example 3: Request with custom trace ID
	fmt.Println("\n=== Example 3: Request with Custom Trace ID ===")
	ctx3, cancel3 := context.WithTimeout(context.Background(), time.Second)
	defer cancel3()

	// Add custom trace ID to context
	ctx3 = context.WithValue(ctx3, app.TraceIDContextKey(), "custom-trace-123")

	resp3, err3 := client.SayHello(ctx3, &pb.HelloRequest{Name: "Zazin"})
	if err3 != nil {
		log.Printf("Error calling SayHello: %v", err3)
	} else {
		fmt.Printf("Response: %s\n", resp3.GetMessage())
	}

	fmt.Println("\nClient examples completed. Check logs for trace information.")
}
