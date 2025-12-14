package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mr1hm/grpc-demo/internal/config"
	"github.com/mr1hm/grpc-demo/internal/gateway"
	"github.com/mr1hm/grpc-demo/internal/user"
	"github.com/mr1hm/grpc-demo/proto/gatewaypb"
	"github.com/mr1hm/grpc-demo/proto/userpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.New(":50051", ":50052")

	// Start User Service (internal)
	go startUserService(cfg)

	// Start Gateway Service (public-facing, calls User service)
	go startGatewayService(cfg)

	cfg.Info("===========================================")
	cfg.Info("gRPC Demo - Inter-service Communication")
	cfg.Info("===========================================")
	cfg.Infof("User Service (internal) running on %s", cfg.UserServicePort)
	cfg.Infof("Gateway Service (public) running on %s", cfg.GatewayServicePort)
	cfg.Info("-------------------------------------------")
	cfg.Info("Test with grpcurl:")
	cfg.Info("  # Register a user via Gateway")
	cfg.Info("  grpcurl -plaintext -d '{\"name\": \"John\", \"email\": \"john@example.com\"}' localhost:50052 gatewaypb.GatewayService/RegisterUser")
	cfg.Info("")
	cfg.Info("  # Get user profile via Gateway (which calls User service internally)")
	cfg.Info("  grpcurl -plaintext -d '{\"user_id\": \"user-1\"}' localhost:50052 gatewaypb.GatewayService/GetUserProfile")
	cfg.Info("===========================================")

	// Wait for interrupt signal.
	// Mainly here to prevent main()
	// from exiting early and allow graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cfg.Info("Shutting down...")
}

func startUserService(cfg *config.Config) {
	lis, err := net.Listen("tcp", cfg.UserServicePort)
	if err != nil {
		cfg.Fatalf("User service failed to listen: %v", err)
	}

	server := grpc.NewServer()
	userpb.RegisterUserServiceServer(server, user.NewService(cfg))
	reflection.Register(server)

	cfg.Infof("[User Service] Starting on %s", cfg.UserServicePort)
	if err := server.Serve(lis); err != nil {
		cfg.Fatalf("User service failed to serve: %v", err)
	}
}

func startGatewayService(cfg *config.Config) {
	lis, err := net.Listen("tcp", cfg.GatewayServicePort)
	if err != nil {
		cfg.Fatalf("Gateway service failed to listen: %v", err)
	}

	gatewaySvc, err := gateway.NewService(cfg, "localhost"+cfg.UserServicePort)
	if err != nil {
		cfg.Fatalf("Failed to create gateway service: %v", err)
	}

	server := grpc.NewServer()
	gatewaypb.RegisterGatewayServiceServer(server, gatewaySvc)
	reflection.Register(server)

	cfg.Infof("[Gateway Service] Starting on %s", cfg.GatewayServicePort)
	if err := server.Serve(lis); err != nil {
		cfg.Fatalf("Gateway service failed to serve: %v", err)
	}
}
