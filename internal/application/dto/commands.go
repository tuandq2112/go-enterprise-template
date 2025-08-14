package dto

// ==================== COMMANDS ====================

// CreateUserCommand represents a command to create a user
type CreateUserCommand struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required,min=2,max=100"`
}

// CreateUserCommandResponse represents the response of creating a user command
type CreateUserCommandResponse struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// UpdateUserCommand represents a command to update a user
type UpdateUserCommand struct {
	UserID string `json:"user_id" validate:"required"`
	Name   string `json:"name" validate:"required,min=2,max=100"`
}

// UpdateUserCommandResponse represents the response of updating a user command
type UpdateUserCommandResponse struct {
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	UpdatedAt string `json:"updated_at"`
}

// DeleteUserCommand represents a command to delete a user
type DeleteUserCommand struct {
	UserID string `json:"user_id" validate:"required"`
}

// DeleteUserCommandResponse represents the response of deleting a user command
type DeleteUserCommandResponse struct {
	UserID    string `json:"user_id"`
	DeletedAt string `json:"deleted_at"`
	Success   bool   `json:"success"`
}
