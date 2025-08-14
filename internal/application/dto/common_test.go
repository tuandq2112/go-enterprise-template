package dto_test

import (
	"testing"

	"go-clean-ddd-es-template/internal/application/dto"

	"github.com/stretchr/testify/assert"
)

func TestPaginationRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.PaginationRequest
		wantErr bool
	}{
		{
			name: "valid pagination request",
			req: dto.PaginationRequest{
				Page:     1,
				PageSize: 10,
			},
			wantErr: false,
		},
		{
			name: "zero page",
			req: dto.PaginationRequest{
				Page:     0,
				PageSize: 10,
			},
			wantErr: true,
		},
		{
			name: "zero page size",
			req: dto.PaginationRequest{
				Page:     1,
				PageSize: 0,
			},
			wantErr: true,
		},
		{
			name: "negative page",
			req: dto.PaginationRequest{
				Page:     -1,
				PageSize: 10,
			},
			wantErr: true,
		},
		{
			name: "negative page size",
			req: dto.PaginationRequest{
				Page:     1,
				PageSize: -10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dto.ValidateRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPaginationRequest_CalculateOffset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		expected int
	}{
		{"first page", 1, 10, 0},
		{"second page", 2, 10, 10},
		{"third page", 3, 5, 10},
		{"large page", 100, 20, 1980},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := dto.PaginationRequest{
				Page:     tt.page,
				PageSize: tt.pageSize,
			}
			offset := (req.Page - 1) * req.PageSize
			assert.Equal(t, tt.expected, offset)
		})
	}
}

func TestNewErrorResponse(t *testing.T) {
	message := "test error message"
	code := "TEST_ERROR"

	resp := dto.NewErrorResponse(message, code)

	assert.Equal(t, message, resp.Error)
	assert.Equal(t, code, resp.Code)
}

func TestNewValidationErrorResponse(t *testing.T) {
	message := "validation failed"
	details := map[string]string{
		"email": "invalid email format",
		"name":  "name is required",
	}

	resp := dto.NewValidationErrorResponse(message, details)

	assert.Equal(t, message, resp.Error)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.Equal(t, details, resp.Details)
}

func TestValidationError_Fields(t *testing.T) {
	ve := dto.ValidationError{
		Field:   "email",
		Message: "invalid email format",
		Value:   "invalid@email",
	}

	assert.Equal(t, "email", ve.Field)
	assert.Equal(t, "invalid email format", ve.Message)
	assert.Equal(t, "invalid@email", ve.Value)
}

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     interface{}
		wantErr bool
	}{
		{
			name: "valid request",
			req: dto.PaginationRequest{
				Page:     1,
				PageSize: 10,
			},
			wantErr: false,
		},
		{
			name: "invalid request",
			req: dto.PaginationRequest{
				Page:     0,
				PageSize: 10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dto.ValidateRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
