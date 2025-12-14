package user

import (
	"context"
	"fmt"
	"sync"

	"github.com/mr1hm/grpc-demo/internal/config"
	"github.com/mr1hm/grpc-demo/proto/userpb"
	"google.golang.org/grpc/codes"
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
