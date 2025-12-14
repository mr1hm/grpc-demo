package user

import (
	"context"
	"testing"

	"github.com/mr1hm/grpc-demo/internal/config"
	"github.com/mr1hm/grpc-demo/proto/userpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestService() *Service {
	cfg := config.New(":50051", ":50052")
	return NewService(cfg)
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Service)
		userID  string
		want    *userpb.GetUserResponse
		wantErr codes.Code
	}{
		{
			name: "user exists",
			setup: func(s *Service) {
				s.users["user-1"] = &userpb.GetUserResponse{
					UserId: "user-1",
					Name:   "John",
					Email:  "john@example.com",
				}
			},
			userID: "user-1",
			want: &userpb.GetUserResponse{
				UserId: "user-1",
				Name:   "John",
				Email:  "john@example.com",
			},
			wantErr: codes.OK,
		},
		{
			name:    "user not found",
			setup:   func(s *Service) {},
			userID:  "nonexistent",
			want:    nil,
			wantErr: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService()
			tt.setup(svc)

			got, err := svc.GetUser(context.Background(), &userpb.GetUserRequest{
				UserId: tt.userID,
			})

			if tt.wantErr != codes.OK {
				if err == nil {
					t.Fatalf("expected error with code %v, got nil", tt.wantErr)
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("expected gRPC status error, got %v", err)
				}
				if st.Code() != tt.wantErr {
					t.Errorf("expected code %v, got %v", tt.wantErr, st.Code())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.UserId != tt.want.UserId || got.Name != tt.want.Name || got.Email != tt.want.Email {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name  string
		input *userpb.CreateUserRequest
		want  *userpb.CreateUserResponse
	}{
		{
			name: "create first user",
			input: &userpb.CreateUserRequest{
				Name:  "Alice",
				Email: "alice@example.com",
			},
			want: &userpb.CreateUserResponse{
				UserId: "user-1",
				Name:   "Alice",
				Email:  "alice@example.com",
			},
		},
		{
			name: "create second user",
			input: &userpb.CreateUserRequest{
				Name:  "Bob",
				Email: "bob@example.com",
			},
			want: &userpb.CreateUserResponse{
				UserId: "user-2",
				Name:   "Bob",
				Email:  "bob@example.com",
			},
		},
	}

	svc := newTestService()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.CreateUser(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.UserId != tt.want.UserId {
				t.Errorf("UserId = %q, want %q", got.UserId, tt.want.UserId)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Email != tt.want.Email {
				t.Errorf("Email = %q, want %q", got.Email, tt.want.Email)
			}
		})
	}
}

func TestCreateUser_ThenGetUser(t *testing.T) {
	svc := newTestService()

	// Create a user
	created, err := svc.CreateUser(context.Background(), &userpb.CreateUserRequest{
		Name:  "Charlie",
		Email: "charlie@example.com",
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Retrieve the same user
	got, err := svc.GetUser(context.Background(), &userpb.GetUserRequest{
		UserId: created.UserId,
	})
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if got.UserId != created.UserId || got.Name != created.Name || got.Email != created.Email {
		t.Errorf("GetUser returned %+v, want matching created user %+v", got, created)
	}
}
