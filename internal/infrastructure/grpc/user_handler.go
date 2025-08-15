package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/application/services"
	"go-clean-ddd-es-template/pkg/tracing"
	"go-clean-ddd-es-template/proto/user"
)

// UserGRPCServer implements the gRPC UserService server
type UserGRPCServer struct {
	user.UnimplementedUserServiceServer
	userService *services.UserService
	tracer      *tracing.Tracer
}

// NewUserGRPCServer creates a new user gRPC server
func NewUserGRPCServer(userService *services.UserService, tracer *tracing.Tracer) *UserGRPCServer {
	return &UserGRPCServer{
		userService: userService,
		tracer:      tracer,
	}
}

// CreateUser implements user.UserServiceServer.CreateUser
func (s *UserGRPCServer) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx, "UserGRPCServer.CreateUser")
	defer span.End()

	cmd := dto.CreateUserCommand{
		Email: req.Email,
		Name:  req.Name,
	}

	if err := dto.ValidateRequest(cmd); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	response, err := s.userService.CreateUser(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &user.CreateUserResponse{
		User: &user.User{
			Id:        response.UserID,
			Email:     response.Email,
			Name:      response.Name,
			CreatedAt: response.CreatedAt,
			UpdatedAt: response.CreatedAt,
		},
	}, nil
}

// GetUser implements user.UserServiceServer.GetUser
func (s *UserGRPCServer) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx, "UserGRPCServer.GetUser")
	defer span.End()

	query := dto.GetUserQuery{
		UserID: req.Id,
	}

	if err := dto.ValidateRequest(query); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	response, err := s.userService.GetUser(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return &user.GetUserResponse{
		User: &user.User{
			Id:        response.UserID,
			Email:     response.Email,
			Name:      response.Name,
			CreatedAt: response.CreatedAt,
			UpdatedAt: response.UpdatedAt,
		},
	}, nil
}

// ListUsers implements user.UserServiceServer.ListUsers
func (s *UserGRPCServer) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx, "UserGRPCServer.ListUsers")
	defer span.End()

	query := dto.ListUsersQuery{
		Page:     1,
		PageSize: 10,
	}

	if err := dto.ValidateRequest(query); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	response, err := s.userService.ListUsers(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	users := make([]*user.User, len(response.Users))
	for i, u := range response.Users {
		users[i] = &user.User{
			Id:        u.UserID,
			Email:     u.Email,
			Name:      u.Name,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.CreatedAt,
		}
	}

	return &user.ListUsersResponse{
		Users: users,
	}, nil
}

// UpdateUser implements user.UserServiceServer.UpdateUser
func (s *UserGRPCServer) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx, "UserGRPCServer.UpdateUser")
	defer span.End()

	cmd := dto.UpdateUserCommand{
		UserID: req.Id,
		Name:   req.Name,
	}

	if err := dto.ValidateRequest(cmd); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	response, err := s.userService.UpdateUser(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &user.UpdateUserResponse{
		User: &user.User{
			Id:        response.UserID,
			Name:      response.Name,
			UpdatedAt: response.UpdatedAt,
		},
	}, nil
}

// DeleteUser implements user.UserServiceServer.DeleteUser
func (s *UserGRPCServer) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx, "UserGRPCServer.DeleteUser")
	defer span.End()

	cmd := dto.DeleteUserCommand{
		UserID: req.Id,
	}

	if err := dto.ValidateRequest(cmd); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	response, err := s.userService.DeleteUser(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &user.DeleteUserResponse{
		Success: response.Success,
	}, nil
}
