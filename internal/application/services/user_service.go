package services

import (
	"context"

	"go-clean-ddd-es-template/internal/application/commands"
	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/application/queries"
)

// UserService combines all command and query handlers for user operations
type UserService struct {
	createCommandHandler   *commands.UserCreateCommandHandler
	updateCommandHandler   *commands.UserUpdateCommandHandler
	deleteCommandHandler   *commands.UserDeleteCommandHandler
	getQueryHandler        *queries.UserGetQueryHandler
	listQueryHandler       *queries.UserListQueryHandler
	getByEmailQueryHandler *queries.UserGetByEmailQueryHandler
	eventsQueryHandler     *queries.UserEventsQueryHandler
}

// NewUserService creates a new user service
func NewUserService(
	createCommandHandler *commands.UserCreateCommandHandler,
	updateCommandHandler *commands.UserUpdateCommandHandler,
	deleteCommandHandler *commands.UserDeleteCommandHandler,
	getQueryHandler *queries.UserGetQueryHandler,
	listQueryHandler *queries.UserListQueryHandler,
	getByEmailQueryHandler *queries.UserGetByEmailQueryHandler,
	eventsQueryHandler *queries.UserEventsQueryHandler,
) *UserService {
	return &UserService{
		createCommandHandler:   createCommandHandler,
		updateCommandHandler:   updateCommandHandler,
		deleteCommandHandler:   deleteCommandHandler,
		getQueryHandler:        getQueryHandler,
		listQueryHandler:       listQueryHandler,
		getByEmailQueryHandler: getByEmailQueryHandler,
		eventsQueryHandler:     eventsQueryHandler,
	}
}

// ==================== COMMANDS ====================

// CreateUser executes the create user command
func (s *UserService) CreateUser(ctx context.Context, cmd dto.CreateUserCommand) (*dto.CreateUserCommandResponse, error) {
	return s.createCommandHandler.Handle(ctx, cmd)
}

// UpdateUser executes the update user command
func (s *UserService) UpdateUser(ctx context.Context, cmd dto.UpdateUserCommand) (*dto.UpdateUserCommandResponse, error) {
	return s.updateCommandHandler.Handle(ctx, cmd)
}

// DeleteUser executes the delete user command
func (s *UserService) DeleteUser(ctx context.Context, cmd dto.DeleteUserCommand) (*dto.DeleteUserCommandResponse, error) {
	return s.deleteCommandHandler.Handle(ctx, cmd)
}

// ==================== QUERIES ====================

// GetUser executes the get user query
func (s *UserService) GetUser(ctx context.Context, query dto.GetUserQuery) (*dto.GetUserQueryResponse, error) {
	return s.getQueryHandler.Handle(ctx, query)
}

// ListUsers executes the list users query
func (s *UserService) ListUsers(ctx context.Context, query dto.ListUsersQuery) (*dto.ListUsersQueryResponse, error) {
	return s.listQueryHandler.Handle(ctx, query)
}

// GetUserByEmail executes the get user by email query
func (s *UserService) GetUserByEmail(ctx context.Context, query dto.GetUserByEmailQuery) (*dto.GetUserByEmailQueryResponse, error) {
	return s.getByEmailQueryHandler.Handle(ctx, query)
}

// GetUserEvents executes the get user events query
func (s *UserService) GetUserEvents(ctx context.Context, query dto.GetUserEventsQuery) (*dto.GetUserEventsQueryResponse, error) {
	return s.eventsQueryHandler.Handle(ctx, query)
}
