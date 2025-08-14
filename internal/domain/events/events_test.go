package events

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		data      interface{}
		version   int
		wantErr   bool
	}{
		{
			name:      "valid event",
			eventType: "user.created",
			data:      map[string]string{"user_id": "123"},
			version:   1,
			wantErr:   false,
		},
		{
			name:      "empty event type",
			eventType: "",
			data:      map[string]string{"user_id": "123"},
			version:   1,
			wantErr:   false,
		},
		{
			name:      "nil data",
			eventType: "user.created",
			data:      nil,
			version:   1,
			wantErr:   false,
		},
		{
			name:      "zero version",
			eventType: "user.created",
			data:      map[string]string{"user_id": "123"},
			version:   0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := NewEvent(tt.eventType, tt.data, tt.version)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, event.ID)
			assert.Equal(t, tt.eventType, event.Type)
			assert.Equal(t, tt.version, event.Version)
			assert.NotZero(t, event.Timestamp)
			assert.NotNil(t, event.Data)
		})
	}
}

func TestNewEvent_WithUserCreatedEvent(t *testing.T) {
	userEvent := &UserCreatedEvent{
		UserID:    "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}

	event, err := NewEvent("user.created", userEvent, 1)
	assert.NoError(t, err)
	assert.Equal(t, "user.created", event.Type)
	assert.Equal(t, 1, event.Version)

	// Verify data can be unmarshaled back to UserCreatedEvent
	var unmarshaledEvent UserCreatedEvent
	err = json.Unmarshal(event.Data, &unmarshaledEvent)
	assert.NoError(t, err)
	assert.Equal(t, userEvent.UserID, unmarshaledEvent.UserID)
	assert.Equal(t, userEvent.Email, unmarshaledEvent.Email)
	assert.Equal(t, userEvent.Name, unmarshaledEvent.Name)
}

func TestNewEvent_WithUserUpdatedEvent(t *testing.T) {
	userEvent := &UserUpdatedEvent{
		UserID:    "user-123",
		Name:      "Updated Name",
		UpdatedAt: time.Now(),
	}

	event, err := NewEvent("user.updated", userEvent, 2)
	assert.NoError(t, err)
	assert.Equal(t, "user.updated", event.Type)
	assert.Equal(t, 2, event.Version)

	// Verify data can be unmarshaled back to UserUpdatedEvent
	var unmarshaledEvent UserUpdatedEvent
	err = json.Unmarshal(event.Data, &unmarshaledEvent)
	assert.NoError(t, err)
	assert.Equal(t, userEvent.UserID, unmarshaledEvent.UserID)
	assert.Equal(t, userEvent.Name, unmarshaledEvent.Name)
}

func TestNewEvent_WithUserDeletedEvent(t *testing.T) {
	userEvent := &UserDeletedEvent{
		UserID:    "user-123",
		DeletedAt: time.Now(),
	}

	event, err := NewEvent("user.deleted", userEvent, 3)
	assert.NoError(t, err)
	assert.Equal(t, "user.deleted", event.Type)
	assert.Equal(t, 3, event.Version)

	// Verify data can be unmarshaled back to UserDeletedEvent
	var unmarshaledEvent UserDeletedEvent
	err = json.Unmarshal(event.Data, &unmarshaledEvent)
	assert.NoError(t, err)
	assert.Equal(t, userEvent.UserID, unmarshaledEvent.UserID)
}

func TestEvent_JSONSerialization(t *testing.T) {
	userEvent := &UserCreatedEvent{
		UserID:    "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}

	event, err := NewEvent("user.created", userEvent, 1)
	assert.NoError(t, err)

	// Test marshaling
	jsonData, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test unmarshaling
	var unmarshaledEvent Event
	err = json.Unmarshal(jsonData, &unmarshaledEvent)
	assert.NoError(t, err)
	assert.Equal(t, event.ID, unmarshaledEvent.ID)
	assert.Equal(t, event.Type, unmarshaledEvent.Type)
	assert.Equal(t, event.Version, unmarshaledEvent.Version)
	assert.Equal(t, event.Data, unmarshaledEvent.Data)
}

func TestGenerateEventID(t *testing.T) {
	id1 := generateEventID()
	id2 := generateEventID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)

	// Test format - should be timestamp format
	assert.Len(t, id1, 14) // YYYYMMDDHHMMSS format
	assert.Len(t, id2, 14)

	// Test that IDs are valid timestamp strings
	_, err1 := time.Parse("20060102150405", id1)
	_, err2 := time.Parse("20060102150405", id2)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
}

func TestUserCreatedEvent_JSONSerialization(t *testing.T) {
	event := &UserCreatedEvent{
		UserID:    "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}

	jsonData, err := json.Marshal(event)
	assert.NoError(t, err)

	var unmarshaledEvent UserCreatedEvent
	err = json.Unmarshal(jsonData, &unmarshaledEvent)
	assert.NoError(t, err)
	assert.Equal(t, event.UserID, unmarshaledEvent.UserID)
	assert.Equal(t, event.Email, unmarshaledEvent.Email)
	assert.Equal(t, event.Name, unmarshaledEvent.Name)
}

func TestUserUpdatedEvent_JSONSerialization(t *testing.T) {
	event := &UserUpdatedEvent{
		UserID:    "user-123",
		Name:      "Updated Name",
		UpdatedAt: time.Now(),
	}

	jsonData, err := json.Marshal(event)
	assert.NoError(t, err)

	var unmarshaledEvent UserUpdatedEvent
	err = json.Unmarshal(jsonData, &unmarshaledEvent)
	assert.NoError(t, err)
	assert.Equal(t, event.UserID, unmarshaledEvent.UserID)
	assert.Equal(t, event.Name, unmarshaledEvent.Name)
}

func TestUserDeletedEvent_JSONSerialization(t *testing.T) {
	event := &UserDeletedEvent{
		UserID:    "user-123",
		DeletedAt: time.Now(),
	}

	jsonData, err := json.Marshal(event)
	assert.NoError(t, err)

	var unmarshaledEvent UserDeletedEvent
	err = json.Unmarshal(jsonData, &unmarshaledEvent)
	assert.NoError(t, err)
	assert.Equal(t, event.UserID, unmarshaledEvent.UserID)
}
