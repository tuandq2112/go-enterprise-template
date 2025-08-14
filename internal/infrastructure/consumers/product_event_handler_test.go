package consumers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductEventHandler_HandleEvent(t *testing.T) {
	tests := []struct {
		name          string
		eventType     string
		eventData     map[string]interface{}
		expectedError bool
	}{
		{
			name:      "successful product.created event handling",
			eventType: "product.created",
			eventData: map[string]interface{}{
				"product_id": "product-123",
				"name":       "Test Product",
				"price":      99.99,
			},
			expectedError: false,
		},
		{
			name:      "successful product.updated event handling",
			eventType: "product.updated",
			eventData: map[string]interface{}{
				"product_id": "product-123",
				"name":       "Updated Product",
				"price":      149.99,
			},
			expectedError: false,
		},
		{
			name:      "successful product.deleted event handling",
			eventType: "product.deleted",
			eventData: map[string]interface{}{
				"product_id": "product-123",
			},
			expectedError: false,
		},
		{
			name:      "unknown event type",
			eventType: "product.unknown",
			eventData: map[string]interface{}{
				"product_id": "product-123",
			},
			expectedError: false, // Should not error, just skip
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			handler := NewProductEventHandler()

			// Handle event
			err := handler.HandleEvent(context.Background(), tt.eventType, tt.eventData)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProductEventHandler_HandleEventWithEmptyData(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		eventData map[string]interface{}
	}{
		{
			name:      "empty event data",
			eventType: "product.created",
			eventData: map[string]interface{}{},
		},
		{
			name:      "nil event data",
			eventType: "product.updated",
			eventData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			handler := NewProductEventHandler()

			// Handle event
			err := handler.HandleEvent(context.Background(), tt.eventType, tt.eventData)

			// Should handle gracefully without error
			assert.NoError(t, err)
		})
	}
}

func TestProductEventHandler_NewProductEventHandler(t *testing.T) {
	// Test constructor
	handler := NewProductEventHandler()

	// Verify handler is created
	assert.NotNil(t, handler)
	assert.IsType(t, &ProductEventHandler{}, handler)
}
