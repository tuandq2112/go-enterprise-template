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

// ==================== AUTH COMMANDS ====================

// RegisterCommand represents a command to register a new user
type RegisterCommand struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterResponse represents the response of register command
type RegisterResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Token  string `json:"token"`
}

// LoginCommand represents a command to login a user
type LoginCommand struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the response of login command
type LoginResponse struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Roles  []string `json:"roles"`
	Token  string   `json:"token"`
}

// ChangePasswordCommand represents a command to change password
type ChangePasswordCommand struct {
	UserID          string `json:"user_id" validate:"required"`
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// ChangePasswordResponse represents the response of change password command
type ChangePasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ValidateTokenResponse represents the response of validate token command
type ValidateTokenResponse struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
}

// RefreshTokenResponse represents the response of refresh token command
type RefreshTokenResponse struct {
	Token string `json:"token"`
}
