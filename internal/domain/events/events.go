package events

import (
	"encoding/json"
	"time"
)

// Event represents a domain event
type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Data      []byte    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	Version   int       `json:"version"`
}

// NewEvent creates a new domain event
func NewEvent(eventType string, data interface{}, version int) (*Event, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Data:      jsonData,
		Timestamp: time.Now(),
		Version:   version,
	}, nil
}

// UserCreatedEvent represents a user creation event
type UserCreatedEvent struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// UserUpdatedEvent represents a user update event
type UserUpdatedEvent struct {
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserDeletedEvent represents a user deletion event
type UserDeletedEvent struct {
	UserID    string    `json:"user_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

func generateEventID() string {
	// This would typically use a UUID generator
	// For now, using a simple timestamp-based ID
	return time.Now().Format("20060102150405")
}
