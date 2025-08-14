package queries

import (
	"context"
	"fmt"

	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/domain/repositories"
)

// UserGetQueryHandler handles the get user query (read operation)
// Uses MongoDB read repository for optimized read performance
type UserGetQueryHandler struct {
	userReadRepository repositories.UserReadRepository
}

// NewUserGetQueryHandler creates a new user get query handler
func NewUserGetQueryHandler(userReadRepository repositories.UserReadRepository) *UserGetQueryHandler {
	return &UserGetQueryHandler{
		userReadRepository: userReadRepository,
	}
}

// Handle handles the get user query
func (h *UserGetQueryHandler) Handle(ctx context.Context, query dto.GetUserQuery) (*dto.GetUserQueryResponse, error) {
	// Get user from MongoDB read model (optimized for queries)
	user, err := h.userReadRepository.GetUserByID(ctx, query.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert to response DTO
	response := &dto.GetUserQueryResponse{
		UserID:    user.UserID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return response, nil
}
