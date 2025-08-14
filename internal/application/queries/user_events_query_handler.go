package queries

import (
	"context"
	"encoding/json"
	"fmt"

	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/domain/repositories"
)

// UserEventsQueryHandler handles the get user events query (read operation)
// Uses MongoDB read repository for optimized read performance
type UserEventsQueryHandler struct {
	userReadRepository repositories.UserReadRepository
}

// NewUserEventsQueryHandler creates a new user events query handler
func NewUserEventsQueryHandler(userReadRepository repositories.UserReadRepository) *UserEventsQueryHandler {
	return &UserEventsQueryHandler{
		userReadRepository: userReadRepository,
	}
}

// Handle handles the get user events query
func (h *UserEventsQueryHandler) Handle(ctx context.Context, query dto.GetUserEventsQuery) (*dto.GetUserEventsQueryResponse, error) {
	// Get events from MongoDB read model (optimized for queries)
	events, err := h.userReadRepository.GetUserEvents(ctx, query.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user events: %w", err)
	}

	// Convert to response DTO
	eventRecords := make([]dto.EventRecord, len(events))
	for i, event := range events {
		// Convert event data to JSON string
		eventDataJSON, err := json.Marshal(event.EventData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal event data: %w", err)
		}

		eventRecords[i] = dto.EventRecord{
			EventID:   event.ID.Hex(),
			EventType: event.EventType,
			Data:      string(eventDataJSON),
			Timestamp: event.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			Version:   event.Version,
		}
	}

	response := &dto.GetUserEventsQueryResponse{
		UserID: query.UserID,
		Events: eventRecords,
	}

	return response, nil
}
