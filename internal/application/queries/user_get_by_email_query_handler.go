package queries

import (
	"context"
	"fmt"

	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/domain/repositories"
)

// UserGetByEmailQueryHandler handles the get user by email query (read operation)
// Uses MongoDB read repository for optimized read performance
type UserGetByEmailQueryHandler struct {
	userReadRepository repositories.UserReadRepository
}

// NewUserGetByEmailQueryHandler creates a new user get by email query handler
func NewUserGetByEmailQueryHandler(userReadRepository repositories.UserReadRepository) *UserGetByEmailQueryHandler {
	return &UserGetByEmailQueryHandler{
		userReadRepository: userReadRepository,
	}
}

// Handle handles the get user by email query
func (h *UserGetByEmailQueryHandler) Handle(ctx context.Context, query dto.GetUserByEmailQuery) (*dto.GetUserByEmailQueryResponse, error) {
	// Get user from MongoDB read model (optimized for queries)
	user, err := h.userReadRepository.GetUserByEmail(ctx, query.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Convert to response DTO
	response := &dto.GetUserByEmailQueryResponse{
		UserID:    user.UserID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return response, nil
}
