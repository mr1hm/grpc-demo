package gateway

import (
	"context"
	"fmt"
	"net"

	"github.com/mr1hm/grpc-demo/internal/config"
	"github.com/mr1hm/grpc-demo/proto/gatewaypb"
	"github.com/mr1hm/grpc-demo/proto/userpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

// Service implements the GatewayService gRPC server
type Service struct {
	cfg *config.Config
	gatewaypb.UnimplementedGatewayServiceServer
	userClient userpb.UserServiceClient
	conn       *grpc.ClientConn
}

// NewService creates a new Gateway service that connects to the User service
func NewService(cfg *config.Config, userServiceAddr string) *Service {
	conn, err := grpc.NewClient(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		cfg.Fatalf("Failed to connect to user service: %v", err)
	}

	return &Service{
		cfg:        cfg,
		userClient: userpb.NewUserServiceClient(conn),
		conn:       conn,
	}
}

// NewServiceWithClient creates a Gateway service with a provided client (for testing)
func NewServiceWithClient(cfg *config.Config, userClient userpb.UserServiceClient) *Service {
	return &Service{
		cfg:        cfg,
		userClient: userClient,
	}
}

// Close closes the client connection
func (s *Service) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

// Start creates a listener, registers the service, and starts serving in a goroutine.
// Returns the server for graceful shutdown.
func (s *Service) Start() *grpc.Server {
	lis, err := net.Listen("tcp", s.cfg.GatewayServicePort)
	if err != nil {
		s.cfg.Fatalf("Gateway service failed to listen: %v", err)
	}

	server := grpc.NewServer()
	gatewaypb.RegisterGatewayServiceServer(server, s)
	reflection.Register(server)

	go func() {
		s.cfg.Infof("[Gateway Service] Starting on %s", s.cfg.GatewayServicePort)
		if err := server.Serve(lis); err != nil {
			s.cfg.Fatalf("Gateway service error: %v", err)
		}
	}()

	return server
}

// GetUserProfile gets a user profile by calling the internal User service
func (s *Service) GetUserProfile(ctx context.Context, req *gatewaypb.GetUserProfileRequest) (*gatewaypb.GetUserProfileResponse, error) {
	s.cfg.Infof("[Gateway] GetUserProfile called for user: %s", req.UserId)

	// Call internal User service
	userResp, err := s.userClient.GetUser(ctx, &userpb.GetUserRequest{
		UserId: req.UserId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user from user service: %w", err)
	}

	s.cfg.Infof("[Gateway] Received user data from User service: %+v", userResp)

	// Gateway adds additional data/processing
	return &gatewaypb.GetUserProfileResponse{
		UserId: userResp.UserId,
		Name:   userResp.Name,
		Email:  userResp.Email,
		Status: "active", // Gateway enriches the response
	}, nil
}

// RegisterUser registers a new user via the internal User service
func (s *Service) RegisterUser(ctx context.Context, req *gatewaypb.RegisterUserRequest) (*gatewaypb.RegisterUserResponse, error) {
	s.cfg.Infof("[Gateway] RegisterUser called: name=%s, email=%s", req.Name, req.Email)

	// Call internal User service
	userResp, err := s.userClient.CreateUser(ctx, &userpb.CreateUserRequest{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user via user service: %w", err)
	}

	s.cfg.Infof("[Gateway] User created via User service: %+v", userResp)

	return &gatewaypb.RegisterUserResponse{
		UserId:  userResp.UserId,
		Message: fmt.Sprintf("User %s registered successfully", userResp.Name),
	}, nil
}
