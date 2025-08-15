package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go-clean-ddd-es-template/proto/user"
)

const (
	grpcServerAddr = "localhost:9091"
)

func TestGRPCCreateUser(t *testing.T) {
	// Skip if gRPC server is not running
	if !isGRPCServerRunning() {
		t.Skip("gRPC server is not running, skipping integration test")
	}

	// Connect to gRPC server
	conn, err := grpc.Dial(grpcServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	// Create gRPC client
	client := user.NewUserServiceClient(conn)

	// Create user request
	req := &user.CreateUserRequest{
		Email: fmt.Sprintf("grpc-test-%d@example.com", time.Now().Unix()),
		Name:  "gRPC Test User",
	}

	// Call gRPC method
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.CreateUser(ctx, req)
	require.NoError(t, err)

	// Verify response
	assert.NotEmpty(t, resp.User.Id)
	assert.Equal(t, req.Email, resp.User.Email)
	assert.Equal(t, req.Name, resp.User.Name)
	assert.NotNil(t, resp.User.CreatedAt)
	assert.NotNil(t, resp.User.UpdatedAt)
}

func TestGRPCListUsers(t *testing.T) {
	// Skip if gRPC server is not running
	if !isGRPCServerRunning() {
		t.Skip("gRPC server is not running, skipping integration test")
	}

	// Connect to gRPC server
	conn, err := grpc.Dial(grpcServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	// Create gRPC client
	client := user.NewUserServiceClient(conn)

	// List users request
	req := &user.ListUsersRequest{}

	// Call gRPC method
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.ListUsers(ctx, req)
	require.NoError(t, err)

	// Verify response
	assert.NotNil(t, resp.Users)
	assert.GreaterOrEqual(t, len(resp.Users), 0)

	// If there are users, check their structure
	for _, user := range resp.Users {
		assert.NotEmpty(t, user.Id)
		assert.NotEmpty(t, user.Email)
		assert.NotEmpty(t, user.Name)
		assert.NotNil(t, user.CreatedAt)
		assert.NotNil(t, user.UpdatedAt)
	}
}

func TestGRPCCreateAndListUsers(t *testing.T) {
	// Skip if gRPC server is not running
	if !isGRPCServerRunning() {
		t.Skip("gRPC server is not running, skipping integration test")
	}

	// Connect to gRPC server
	conn, err := grpc.Dial(grpcServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	// Create gRPC client
	client := user.NewUserServiceClient(conn)

	// Create a new user
	createReq := &user.CreateUserRequest{
		Email: fmt.Sprintf("grpc-integration-%d@example.com", time.Now().Unix()),
		Name:  "gRPC Integration Test User",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	createResp, err := client.CreateUser(ctx, createReq)
	require.NoError(t, err)

	// List users and verify the new user is there
	listReq := &user.ListUsersRequest{}
	listResp, err := client.ListUsers(ctx, listReq)
	require.NoError(t, err)

	// Find the created user in the list
	found := false
	for _, u := range listResp.Users {
		if u.Id == createResp.User.Id {
			assert.Equal(t, createReq.Email, u.Email)
			assert.Equal(t, createReq.Name, u.Name)
			found = true
			break
		}
	}

	assert.True(t, found, "Created user should be found in the list")
}

func TestGRPCInvalidRequest(t *testing.T) {
	// Skip if gRPC server is not running
	if !isGRPCServerRunning() {
		t.Skip("gRPC server is not running, skipping integration test")
	}

	// Connect to gRPC server
	conn, err := grpc.Dial(grpcServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	// Create gRPC client
	client := user.NewUserServiceClient(conn)

	// Create invalid request (empty email)
	req := &user.CreateUserRequest{
		Email: "", // Invalid: empty email
		Name:  "Test User",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.CreateUser(ctx, req)
	// Should return an error for invalid request
	assert.Error(t, err)
}

// Helper function to check if gRPC server is running
func isGRPCServerRunning() bool {
	conn, err := grpc.Dial(grpcServerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(2*time.Second),
	)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
