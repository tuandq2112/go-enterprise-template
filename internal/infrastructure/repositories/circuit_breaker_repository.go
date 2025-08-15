package repositories

import (
	"context"

	"go-clean-ddd-es-template/internal/domain/entities"
	"go-clean-ddd-es-template/internal/domain/repositories"
	"go-clean-ddd-es-template/pkg/resilience"
)

// CircuitBreakerUserWriteRepository wraps UserWriteRepository with circuit breaker
type CircuitBreakerUserWriteRepository struct {
	repository     repositories.UserWriteRepository
	circuitBreaker *resilience.CircuitBreaker
}

// NewCircuitBreakerUserWriteRepository creates a new circuit breaker repository
func NewCircuitBreakerUserWriteRepository(repository repositories.UserWriteRepository, config resilience.CircuitBreakerConfig) *CircuitBreakerUserWriteRepository {
	return &CircuitBreakerUserWriteRepository{
		repository:     repository,
		circuitBreaker: resilience.NewCircuitBreaker(config),
	}
}

// Create wraps repository.Create with circuit breaker
func (r *CircuitBreakerUserWriteRepository) Create(ctx context.Context, user *entities.User) error {
	_, err := r.circuitBreaker.ExecuteWithResult(ctx, func() (interface{}, error) {
		return nil, r.repository.Create(ctx, user)
	})
	return err
}

// GetByID wraps repository.GetByID with circuit breaker
func (r *CircuitBreakerUserWriteRepository) GetByID(ctx context.Context, userID string) (*entities.User, error) {
	result, err := r.circuitBreaker.ExecuteWithResult(ctx, func() (interface{}, error) {
		return r.repository.GetByID(ctx, userID)
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.User), nil
}

// GetByEmail wraps repository.GetByEmail with circuit breaker
func (r *CircuitBreakerUserWriteRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	result, err := r.circuitBreaker.ExecuteWithResult(ctx, func() (interface{}, error) {
		return r.repository.GetByEmail(ctx, email)
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.User), nil
}

// Update wraps repository.Update with circuit breaker
func (r *CircuitBreakerUserWriteRepository) Update(ctx context.Context, user *entities.User) error {
	_, err := r.circuitBreaker.ExecuteWithResult(ctx, func() (interface{}, error) {
		return nil, r.repository.Update(ctx, user)
	})
	return err
}

// Delete wraps repository.Delete with circuit breaker
func (r *CircuitBreakerUserWriteRepository) Delete(ctx context.Context, userID string) error {
	_, err := r.circuitBreaker.ExecuteWithResult(ctx, func() (interface{}, error) {
		return nil, r.repository.Delete(ctx, userID)
	})
	return err
}

// GetStats returns circuit breaker statistics
func (r *CircuitBreakerUserWriteRepository) GetStats() resilience.CircuitBreakerStats {
	return r.circuitBreaker.GetStats()
}

// ForceOpen forces the circuit breaker to open state
func (r *CircuitBreakerUserWriteRepository) ForceOpen() {
	r.circuitBreaker.ForceOpen()
}

// ForceClose forces the circuit breaker to closed state
func (r *CircuitBreakerUserWriteRepository) ForceClose() {
	r.circuitBreaker.ForceClose()
}

// CircuitBreakerUserReadRepository wraps UserReadRepository with circuit breaker
type CircuitBreakerUserReadRepository struct {
	repository     repositories.UserReadRepository
	circuitBreaker *resilience.CircuitBreaker
}

// NewCircuitBreakerUserReadRepository creates a new circuit breaker read repository
func NewCircuitBreakerUserReadRepository(repository repositories.UserReadRepository, config resilience.CircuitBreakerConfig) *CircuitBreakerUserReadRepository {
	return &CircuitBreakerUserReadRepository{
		repository:     repository,
		circuitBreaker: resilience.NewCircuitBreaker(config),
	}
}

// GetByID wraps repository.GetByID with circuit breaker
func (r *CircuitBreakerUserReadRepository) GetByID(ctx context.Context, userID string) (*entities.UserReadModel, error) {
	result, err := r.circuitBreaker.ExecuteWithResult(ctx, func() (interface{}, error) {
		return r.repository.GetUserByID(ctx, userID)
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.UserReadModel), nil
}

// GetByEmail wraps repository.GetByEmail with circuit breaker
func (r *CircuitBreakerUserReadRepository) GetByEmail(ctx context.Context, email string) (*entities.UserReadModel, error) {
	result, err := r.circuitBreaker.ExecuteWithResult(ctx, func() (interface{}, error) {
		return r.repository.GetUserByEmail(ctx, email)
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.UserReadModel), nil
}

// List wraps repository.List with circuit breaker
func (r *CircuitBreakerUserReadRepository) List(ctx context.Context, limit, offset int) ([]*entities.UserReadModel, error) {
	result, err := r.circuitBreaker.ExecuteWithResult(ctx, func() (interface{}, error) {
		// Convert limit/offset to page/pageSize
		page := (offset / limit) + 1
		if offset == 0 {
			page = 1
		}
		users, _, err := r.repository.ListUsers(ctx, page, limit)
		return users, err
	})
	if err != nil {
		return nil, err
	}
	return result.([]*entities.UserReadModel), nil
}

// GetEventsByType wraps repository.GetEventsByType with circuit breaker
func (r *CircuitBreakerUserReadRepository) GetEventsByType(ctx context.Context, eventType string) ([]*entities.UserEvent, error) {
	result, err := r.circuitBreaker.ExecuteWithResult(ctx, func() (interface{}, error) {
		return r.repository.GetEventsByType(ctx, eventType)
	})
	if err != nil {
		return nil, err
	}
	return result.([]*entities.UserEvent), nil
}

// GetStats returns circuit breaker statistics
func (r *CircuitBreakerUserReadRepository) GetStats() resilience.CircuitBreakerStats {
	return r.circuitBreaker.GetStats()
}

// ForceOpen forces the circuit breaker to open state
func (r *CircuitBreakerUserReadRepository) ForceOpen() {
	r.circuitBreaker.ForceOpen()
}

// ForceClose forces the circuit breaker to closed state
func (r *CircuitBreakerUserReadRepository) ForceClose() {
	r.circuitBreaker.ForceClose()
}
