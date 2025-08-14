package consumers

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ProductEventHandler handles product-specific events
// This demonstrates how to extend the system for other models
type ProductEventHandler struct {
	// In a real implementation, you would inject a ProductReadRepository
	// productReadRepository repositories.ProductReadRepository
}

// NewProductEventHandler creates a new product event handler
func NewProductEventHandler() *ProductEventHandler {
	return &ProductEventHandler{}
}

// HandleEvent handles product events
func (h *ProductEventHandler) HandleEvent(ctx context.Context, eventType string, eventData map[string]interface{}) error {
	switch eventType {
	case "product.created":
		return h.handleProductCreated(ctx, eventData)
	case "product.updated":
		return h.handleProductUpdated(ctx, eventData)
	case "product.deleted":
		return h.handleProductDeleted(ctx, eventData)
	default:
		return fmt.Errorf("unknown product event type: %s", eventType)
	}
}

// handleProductCreated handles product.created event
func (h *ProductEventHandler) handleProductCreated(ctx context.Context, data map[string]interface{}) error {
	productID, _ := data["product_id"].(string)
	name, _ := data["name"].(string)
	price, _ := data["price"].(float64)
	createdAtStr, _ := data["created_at"].(string)

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		createdAt = time.Now()
	}

	// In a real implementation, you would:
	// 1. Create ProductReadModel
	// 2. Save to MongoDB using ProductReadRepository
	// 3. Save event to MongoDB

	log.Printf("Product created: ID=%s, Name=%s, Price=%.2f, CreatedAt=%s",
		productID, name, price, createdAt.Format(time.RFC3339))

	return nil
}

// handleProductUpdated handles product.updated event
func (h *ProductEventHandler) handleProductUpdated(ctx context.Context, data map[string]interface{}) error {
	productID, _ := data["product_id"].(string)
	name, _ := data["name"].(string)
	price, _ := data["price"].(float64)
	updatedAtStr, _ := data["updated_at"].(string)

	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		updatedAt = time.Now()
	}

	// In a real implementation, you would:
	// 1. Get existing ProductReadModel from MongoDB
	// 2. Update the model
	// 3. Save back to MongoDB
	// 4. Save event to MongoDB

	log.Printf("Product updated: ID=%s, Name=%s, Price=%.2f, UpdatedAt=%s",
		productID, name, price, updatedAt.Format(time.RFC3339))

	return nil
}

// handleProductDeleted handles product.deleted event
func (h *ProductEventHandler) handleProductDeleted(ctx context.Context, data map[string]interface{}) error {
	productID, _ := data["product_id"].(string)
	deletedAtStr, _ := data["deleted_at"].(string)

	deletedAt, err := time.Parse(time.RFC3339, deletedAtStr)
	if err != nil {
		deletedAt = time.Now()
	}

	// In a real implementation, you would:
	// 1. Get existing ProductReadModel from MongoDB
	// 2. Soft delete the model
	// 3. Save back to MongoDB
	// 4. Save event to MongoDB

	log.Printf("Product deleted: ID=%s, DeletedAt=%s",
		productID, deletedAt.Format(time.RFC3339))

	return nil
}
