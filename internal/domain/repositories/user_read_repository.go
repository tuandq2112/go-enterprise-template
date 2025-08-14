package repositories

import (
	"context"

	"go-clean-ddd-es-template/internal/domain/entities"
)

// UserReadRepository defines the interface for user read operations (queries)
// This is used for read operations that query optimized read models
type UserReadRepository interface {
	// User read operations
	SaveUser(ctx context.Context, user *entities.UserReadModel) error
	GetUserByID(ctx context.Context, userID string) (*entities.UserReadModel, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.UserReadModel, error)
	ListUsers(ctx context.Context, page, pageSize int) ([]*entities.UserReadModel, int64, error)
	UpdateUser(ctx context.Context, user *entities.UserReadModel) error
	DeleteUser(ctx context.Context, userID string) error

	// Event read operations
	SaveEvent(ctx context.Context, event *entities.UserEvent) error
	GetUserEvents(ctx context.Context, userID string) ([]*entities.UserEvent, error)
	GetEventsByType(ctx context.Context, eventType string) ([]*entities.UserEvent, error)
}
