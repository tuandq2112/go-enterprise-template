package repositories

import (
	"context"

	"go-clean-ddd-es-template/internal/domain/entities"
)

// UserWriteRepository defines the interface for user write operations (commands)
// This is used for write operations that modify state
type UserWriteRepository interface {
	// Write operations
	Create(ctx context.Context, user *entities.User) error
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, userID string) error

	// Read operations for write side (needed for business logic)
	GetByID(ctx context.Context, userID string) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	List(ctx context.Context) ([]*entities.User, error)
}
