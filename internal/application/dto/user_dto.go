package dto

import (
	"go-clean-ddd-es-template/internal/domain/entities"
)

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required,min=2,max=100"`
}

// CreateUserResponse represents a response from creating a user
type CreateUserResponse struct {
	User *entities.User `json:"user"`
}

// GetUserRequest represents a request to get a user
type GetUserRequest struct {
	ID string `json:"id" validate:"required"`
}

// GetUserResponse represents a response from getting a user
type GetUserResponse struct {
	User *entities.User `json:"user"`
}

// ListUsersRequest represents a request to list users
type ListUsersRequest struct {
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"page_size" validate:"min=1,max=100"`
}

// ListUsersResponse represents a response from listing users
type ListUsersResponse struct {
	Users    []*entities.User `json:"users"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required,min=2,max=100"`
}

// UpdateUserResponse represents a response from updating a user
type UpdateUserResponse struct {
	User *entities.User `json:"user"`
}

// DeleteUserRequest represents a request to delete a user
type DeleteUserRequest struct {
	ID string `json:"id" validate:"required"`
}

// DeleteUserResponse represents a response from deleting a user
type DeleteUserResponse struct {
	Success bool `json:"success"`
}

// GetUserEventsRequest represents a request to get user events
type GetUserEventsRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

// GetUserEventsResponse represents a response from getting user events
type GetUserEventsResponse struct {
	Events []interface{} `json:"events"`
}
