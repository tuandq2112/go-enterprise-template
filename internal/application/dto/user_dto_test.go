package dto_test

import (
	"testing"

	"go-clean-ddd-es-template/internal/application/dto"
	"go-clean-ddd-es-template/internal/domain/entities"

	"github.com/stretchr/testify/assert"
)

func TestCreateUserRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.CreateUserRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: dto.CreateUserRequest{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			req: dto.CreateUserRequest{
				Name:  "",
				Email: "john@example.com",
			},
			wantErr: true,
		},
		{
			name: "empty email",
			req: dto.CreateUserRequest{
				Name:  "John Doe",
				Email: "",
			},
			wantErr: true,
		},
		{
			name: "invalid email format",
			req: dto.CreateUserRequest{
				Name:  "John Doe",
				Email: "invalid-email",
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

func TestUpdateUserRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.UpdateUserRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: dto.UpdateUserRequest{
				ID:   "user-123",
				Name: "John Doe",
			},
			wantErr: false,
		},
		{
			name: "empty id",
			req: dto.UpdateUserRequest{
				ID:   "",
				Name: "John Doe",
			},
			wantErr: true,
		},
		{
			name: "empty name",
			req: dto.UpdateUserRequest{
				ID:   "user-123",
				Name: "",
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

func TestGetUserRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.GetUserRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: dto.GetUserRequest{
				ID: "user-123",
			},
			wantErr: false,
		},
		{
			name: "empty id",
			req: dto.GetUserRequest{
				ID: "",
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

func TestDeleteUserRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.DeleteUserRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: dto.DeleteUserRequest{
				ID: "user-123",
			},
			wantErr: false,
		},
		{
			name: "empty id",
			req: dto.DeleteUserRequest{
				ID: "",
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

func TestListUsersRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.ListUsersRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: dto.ListUsersRequest{
				Page:     1,
				PageSize: 10,
			},
			wantErr: false,
		},
		{
			name: "invalid pagination",
			req: dto.ListUsersRequest{
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

func TestGetUserEventsRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.GetUserEventsRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: dto.GetUserEventsRequest{
				UserID: "user-123",
			},
			wantErr: false,
		},
		{
			name: "empty user id",
			req: dto.GetUserEventsRequest{
				UserID: "",
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

func TestListUsersResponse_Fields(t *testing.T) {
	resp := dto.ListUsersResponse{
		Users:    []*entities.User{},
		Total:    2,
		Page:     1,
		PageSize: 10,
	}

	assert.Len(t, resp.Users, 0)
	assert.Equal(t, int64(2), resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 10, resp.PageSize)
}

func TestDeleteUserResponse_Fields(t *testing.T) {
	resp := dto.DeleteUserResponse{
		Success: true,
	}

	assert.True(t, resp.Success)
}
