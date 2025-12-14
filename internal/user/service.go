package user

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/mr1hm/grpc-demo/internal/config"
	"github.com/mr1hm/grpc-demo/proto/userpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Service implements the UserService gRPC server
type Service struct {
	userpb.UnimplementedUserServiceServer
	cfg    *config.Config
	mu     sync.RWMutex
	users  map[string]*userpb.GetUserResponse
	nextID int
}

// NewService creates a new User service instance
func NewService(cfg *config.Config) *Service {
	return &Service{
		cfg:    cfg,
		users:  make(map[string]*userpb.GetUserResponse),
		nextID: 1,
	}
}

// Start creates a listener, registers the service, and starts serving in a goroutine.
// Returns the server for graceful shutdown.
func (s *Service) Start() *grpc.Server {
	lis, err := net.Listen("tcp", s.cfg.UserServicePort)
	if err != nil {
		s.cfg.Fatalf("User service failed to listen: %v", err)
	}

	server := grpc.NewServer()
	userpb.RegisterUserServiceServer(server, s)
	reflection.Register(server)

	go func() {
		s.cfg.Infof("[User Service] Starting on %s", s.cfg.UserServicePort)
		if err := server.Serve(lis); err != nil {
			s.cfg.Fatalf("User service error: %v", err)
		}
	}()

	return server
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	s.cfg.Infof("[User] GetUser called with ID: %s", req.UserId)
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[req.UserId]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "user %s not found", req.UserId)
	}

	return user, nil
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	s.cfg.Infof("[User] CreateUser called with Name: %s - Email: %s", req.Name, req.Email)
	s.mu.Lock()
	defer s.mu.Unlock()

	userID := fmt.Sprintf("user-%d", s.nextID)
	s.nextID++

	user := &userpb.GetUserResponse{
		UserId: userID,
		Name:   req.Name,
		Email:  req.Email,
	}
	s.users[userID] = user

	return &userpb.CreateUserResponse{
		UserId: userID,
		Name:   req.Name,
		Email:  req.Email,
	}, nil
}
