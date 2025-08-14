package queries

import (
	"context"
	"fmt"

	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/domain/repositories"
)

// UserListQueryHandler handles the list users query (read operation)
// Uses MongoDB read repository for optimized read performance
type UserListQueryHandler struct {
	userReadRepository repositories.UserReadRepository
}

// NewUserListQueryHandler creates a new user list query handler
func NewUserListQueryHandler(userReadRepository repositories.UserReadRepository) *UserListQueryHandler {
	return &UserListQueryHandler{
		userReadRepository: userReadRepository,
	}
}

// Handle handles the list users query
func (h *UserListQueryHandler) Handle(ctx context.Context, query dto.ListUsersQuery) (*dto.ListUsersQueryResponse, error) {
	// Get users from MongoDB read model (optimized for queries)
	users, total, err := h.userReadRepository.ListUsers(ctx, query.Page, query.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Convert to response DTO
	userSummaries := make([]dto.UserSummary, len(users))
	for i, user := range users {
		userSummaries[i] = dto.UserSummary{
			UserID:    user.UserID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response := &dto.ListUsersQueryResponse{
		Users:    userSummaries,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}

	return response, nil
}
