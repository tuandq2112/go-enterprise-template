package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go-clean-ddd-es-template/internal/domain/entities"
	"go-clean-ddd-es-template/internal/infrastructure/database"
)

// PostgresUserWriteRepository implements UserWriteRepository using PostgreSQL
type PostgresUserWriteRepository struct {
	db database.Database
}

// NewPostgresUserWriteRepository creates a new PostgreSQL user write repository
func NewPostgresUserWriteRepository(db database.Database) *PostgresUserWriteRepository {
	return &PostgresUserWriteRepository{
		db: db,
	}
}

// Create saves a new user to PostgreSQL (write database)
func (r *PostgresUserWriteRepository) Create(ctx context.Context, user *entities.User) error {
	// Get underlying database connection
	dbConn := r.db.GetDB()
	if dbConn == nil {
		return errors.New("database connection not available")
	}

	// Cast to sql.DB
	sqlDB, ok := dbConn.(*sql.DB)
	if !ok {
		return errors.New("invalid database connection type - expected sql.DB")
	}

	// Generate user ID if not set
	if user.ID.IsZero() {
		user.ID = entities.NewUserID()
	}

	// Set timestamps
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = time.Now()

	// Insert user using raw SQL
	query := `
		INSERT INTO users (id, email, name, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := sqlDB.ExecContext(ctx, query,
		user.GetID(),
		user.GetEmail(),
		user.GetName(),
		user.GetPasswordHash(),
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID from PostgreSQL (for write operations)
func (r *PostgresUserWriteRepository) GetByID(ctx context.Context, userID string) (*entities.User, error) {
	// Get underlying database connection
	dbConn := r.db.GetDB()
	if dbConn == nil {
		return nil, errors.New("database connection not available")
	}

	// Cast to sql.DB
	sqlDB, ok := dbConn.(*sql.DB)
	if !ok {
		return nil, errors.New("invalid database connection type - expected sql.DB")
	}

	query := `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var id, email, name, passwordHash string
	var createdAt, updatedAt time.Time

	err := sqlDB.QueryRowContext(ctx, query, userID).Scan(
		&id, &email, &name, &passwordHash, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	// Create user entity
	user, err := entities.NewUser(email, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create user entity: %w", err)
	}

	// Set additional fields
	user.SetPasswordHash(passwordHash)
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	return user, nil
}

// GetByEmail retrieves a user by email from PostgreSQL (for write operations)
func (r *PostgresUserWriteRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	// Get underlying database connection
	dbConn := r.db.GetDB()
	if dbConn == nil {
		return nil, errors.New("database connection not available")
	}

	// Cast to sql.DB
	sqlDB, ok := dbConn.(*sql.DB)
	if !ok {
		return nil, errors.New("invalid database connection type - expected sql.DB")
	}

	query := `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var id, userEmail, name, passwordHash string
	var createdAt, updatedAt time.Time

	err := sqlDB.QueryRowContext(ctx, query, email).Scan(
		&id, &userEmail, &name, &passwordHash, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Create user entity
	user, err := entities.NewUser(userEmail, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create user entity: %w", err)
	}

	// Set additional fields
	user.SetPasswordHash(passwordHash)
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	return user, nil
}

// Update updates an existing user in PostgreSQL
func (r *PostgresUserWriteRepository) Update(ctx context.Context, user *entities.User) error {
	// Get underlying database connection
	dbConn := r.db.GetDB()
	if dbConn == nil {
		return errors.New("database connection not available")
	}

	// Cast to sql.DB
	sqlDB, ok := dbConn.(*sql.DB)
	if !ok {
		return errors.New("invalid database connection type - expected sql.DB")
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET email = $1, name = $2, password_hash = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`

	result, err := sqlDB.ExecContext(ctx, query,
		user.GetEmail(),
		user.GetName(),
		user.GetPasswordHash(),
		user.UpdatedAt,
		user.GetID(),
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete removes a user from PostgreSQL
func (r *PostgresUserWriteRepository) Delete(ctx context.Context, userID string) error {
	// Get underlying database connection
	dbConn := r.db.GetDB()
	if dbConn == nil {
		return errors.New("database connection not available")
	}

	// Cast to sql.DB
	sqlDB, ok := dbConn.(*sql.DB)
	if !ok {
		return errors.New("invalid database connection type - expected sql.DB")
	}

	query := `
		UPDATE users
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := sqlDB.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// List retrieves all users from PostgreSQL (for write operations)
func (r *PostgresUserWriteRepository) List(ctx context.Context) ([]*entities.User, error) {
	// Get underlying database connection
	dbConn := r.db.GetDB()
	if dbConn == nil {
		return nil, errors.New("database connection not available")
	}

	// In a real implementation, you would query PostgreSQL
	// For now, return a placeholder error
	return nil, fmt.Errorf("PostgreSQL implementation not available - use a real database driver")
}
