package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mr1hm/grpc-demo/internal/config"
	"github.com/mr1hm/grpc-demo/internal/gateway"
	"github.com/mr1hm/grpc-demo/internal/user"
)

func main() {
	cfg := config.New(":50051", ":50052")

	// Start User Service
	userServer := user.NewService(cfg).Start()

	// Start Gateway Service (connects to User service internally)
	gatewaySvc := gateway.NewService(cfg, "localhost"+cfg.UserServicePort)
	gatewayServer := gatewaySvc.Start()

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

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cfg.Info("Gracefully shutting down...")
	userServer.GracefulStop()
	gatewayServer.GracefulStop()
	gatewaySvc.Close()
}
