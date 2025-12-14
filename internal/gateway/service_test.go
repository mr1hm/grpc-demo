package gateway

import (
	"context"
	"errors"
	"testing"

	"github.com/mr1hm/grpc-demo/internal/config"
	"github.com/mr1hm/grpc-demo/proto/gatewaypb"
	"github.com/mr1hm/grpc-demo/proto/userpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockUserClient implements userpb.UserServiceClient for testing
type mockUserClient struct {
	getUser    func(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error)
	createUser func(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error)
}

func (m *mockUserClient) GetUser(ctx context.Context, req *userpb.GetUserRequest, opts ...grpc.CallOption) (*userpb.GetUserResponse, error) {
	return m.getUser(ctx, req)
}

func (m *mockUserClient) CreateUser(ctx context.Context, req *userpb.CreateUserRequest, opts ...grpc.CallOption) (*userpb.CreateUserResponse, error) {
	return m.createUser(ctx, req)
}

func newTestGatewayService(mock *mockUserClient) *Service {
	cfg := config.New(":50051", ":50052")
	svc, _ := NewService(cfg, mock)
	return svc
}

func TestGetUserProfile(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		mockResp   *userpb.GetUserResponse
		mockErr    error
		wantStatus string
		wantErr    bool
	}{
		{
			name:   "success - adds active status",
			userID: "user-1",
			mockResp: &userpb.GetUserResponse{
				UserId: "user-1",
				Name:   "John",
				Email:  "john@example.com",
			},
			wantStatus: "active",
			wantErr:    false,
		},
		{
			name:    "user service returns not found",
			userID:  "nonexistent",
			mockErr: status.Error(codes.NotFound, "user not found"),
			wantErr: true,
		},
		{
			name:    "user service returns internal error",
			userID:  "user-1",
			mockErr: errors.New("connection refused"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockUserClient{
				getUser: func(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
					if req.UserId != tt.userID {
						t.Errorf("GetUser called with wrong userID: got %q, want %q", req.UserId, tt.userID)
					}
					return tt.mockResp, tt.mockErr
				},
			}

			svc := newTestGatewayService(mock)
			got, err := svc.GetUserProfile(context.Background(), &gatewaypb.GetUserProfileRequest{
				UserId: tt.userID,
			})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", got.Status, tt.wantStatus)
			}
			if got.UserId != tt.mockResp.UserId {
				t.Errorf("UserId = %q, want %q", got.UserId, tt.mockResp.UserId)
			}
			if got.Name != tt.mockResp.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.mockResp.Name)
			}
			if got.Email != tt.mockResp.Email {
				t.Errorf("Email = %q, want %q", got.Email, tt.mockResp.Email)
			}
		})
	}
}

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name     string
		input    *gatewaypb.RegisterUserRequest
		mockResp *userpb.CreateUserResponse
		mockErr  error
		wantErr  bool
	}{
		{
			name: "success",
			input: &gatewaypb.RegisterUserRequest{
				Name:  "Alice",
				Email: "alice@example.com",
			},
			mockResp: &userpb.CreateUserResponse{
				UserId: "user-1",
				Name:   "Alice",
				Email:  "alice@example.com",
			},
			wantErr: false,
		},
		{
			name: "user service error",
			input: &gatewaypb.RegisterUserRequest{
				Name:  "Bob",
				Email: "bob@example.com",
			},
			mockErr: errors.New("database error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockUserClient{
				createUser: func(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
					if req.Name != tt.input.Name {
						t.Errorf("CreateUser called with wrong name: got %q, want %q", req.Name, tt.input.Name)
					}
					if req.Email != tt.input.Email {
						t.Errorf("CreateUser called with wrong email: got %q, want %q", req.Email, tt.input.Email)
					}
					return tt.mockResp, tt.mockErr
				},
			}

			svc := newTestGatewayService(mock)
			got, err := svc.RegisterUser(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.UserId != tt.mockResp.UserId {
				t.Errorf("UserId = %q, want %q", got.UserId, tt.mockResp.UserId)
			}
			if got.Message == "" {
				t.Error("Message should not be empty")
			}
		})
	}
}
