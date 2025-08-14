package dto

// ==================== QUERIES ====================

// GetUserQuery represents a query to get a user by ID
type GetUserQuery struct {
	UserID string `json:"user_id" validate:"required"`
}

// GetUserQueryResponse represents the response of getting a user query
type GetUserQueryResponse struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListUsersQuery represents a query to list users with pagination
type ListUsersQuery struct {
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"page_size" validate:"min=1,max=100"`
}

// ListUsersQueryResponse represents the response of listing users query
type ListUsersQueryResponse struct {
	Users    []UserSummary `json:"users"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// UserSummary represents a summary of user data for listing
type UserSummary struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// GetUserByEmailQuery represents a query to get a user by email
type GetUserByEmailQuery struct {
	Email string `json:"email" validate:"required,email"`
}

// GetUserByEmailQueryResponse represents the response of getting a user by email query
type GetUserByEmailQueryResponse struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// GetUserEventsQuery represents a query to get user events
type GetUserEventsQuery struct {
	UserID string `json:"user_id" validate:"required"`
}

// GetUserEventsQueryResponse represents the response of getting user events query
type GetUserEventsQueryResponse struct {
	UserID string        `json:"user_id"`
	Events []EventRecord `json:"events"`
}

// EventRecord represents an event record
type EventRecord struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	Data      string `json:"data"`
	Timestamp string `json:"timestamp"`
	Version   int    `json:"version"`
}
