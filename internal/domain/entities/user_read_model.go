package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserReadModel represents the read model for user stored in MongoDB
type UserReadModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	Email     string             `bson:"email" json:"email"`
	Name      string             `bson:"name" json:"name"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
	Version   int                `bson:"version" json:"version"`
}

// UserEvent represents a user event for serialization (without MongoDB ObjectID)
type UserEvent struct {
	UserID    string                 `json:"user_id"`
	EventType string                 `json:"event_type"`
	EventData map[string]interface{} `json:"event_data"`
	Timestamp time.Time              `json:"timestamp"`
	Version   int                    `json:"version"`
}

// UserSummary represents a summary of user for listing
type UserSummary struct {
	UserID    string    `bson:"user_id" json:"user_id"`
	Email     string    `bson:"email" json:"email"`
	Name      string    `bson:"name" json:"name"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
