package health_test

import (
	"context"
	"testing"
	"time"

	"go-clean-ddd-es-template/pkg/health"

	"github.com/stretchr/testify/assert"
)

func TestNewHealthService(t *testing.T) {
	service := health.NewHealthService()
	assert.NotNil(t, service)
	assert.Len(t, service.Check(context.Background()), 0)
}

func TestHealthService_AddCheck(t *testing.T) {
	service := health.NewHealthService()

	// Add a custom health check
	customCheck := func(ctx context.Context) health.Check {
		return health.Check{
			Name:    "custom",
			Status:  health.StatusHealthy,
			Message: "Custom check passed",
		}
	}

	service.AddCheck(customCheck)

	checks := service.Check(context.Background())
	assert.Len(t, checks, 1)
	assert.Equal(t, "custom", checks[0].Name)
	assert.Equal(t, health.StatusHealthy, checks[0].Status)
	assert.Equal(t, "Custom check passed", checks[0].Message)
}

func TestHealthService_OverallStatus(t *testing.T) {
	service := health.NewHealthService()

	tests := []struct {
		name     string
		checks   []health.Check
		expected health.Status
	}{
		{
			name:     "no checks",
			checks:   []health.Check{},
			expected: health.StatusHealthy,
		},
		{
			name: "all healthy",
			checks: []health.Check{
				{Name: "check1", Status: health.StatusHealthy},
				{Name: "check2", Status: health.StatusHealthy},
			},
			expected: health.StatusHealthy,
		},
		{
			name: "one degraded",
			checks: []health.Check{
				{Name: "check1", Status: health.StatusHealthy},
				{Name: "check2", Status: health.StatusDegraded},
			},
			expected: health.StatusDegraded,
		},
		{
			name: "one unhealthy",
			checks: []health.Check{
				{Name: "check1", Status: health.StatusHealthy},
				{Name: "check2", Status: health.StatusUnhealthy},
			},
			expected: health.StatusUnhealthy,
		},
		{
			name: "degraded and unhealthy",
			checks: []health.Check{
				{Name: "check1", Status: health.StatusDegraded},
				{Name: "check2", Status: health.StatusUnhealthy},
			},
			expected: health.StatusUnhealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.OverallStatus(tt.checks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheck_Fields(t *testing.T) {
	check := health.Check{
		Name:     "test",
		Status:   health.StatusHealthy,
		Message:  "Test check passed",
		Duration: 100 * time.Millisecond,
		Details: map[string]interface{}{
			"key": "value",
		},
	}

	assert.Equal(t, "test", check.Name)
	assert.Equal(t, health.StatusHealthy, check.Status)
	assert.Equal(t, "Test check passed", check.Message)
	assert.Equal(t, 100*time.Millisecond, check.Duration)
	assert.Equal(t, "value", check.Details["key"])
}

func TestSystemCheck(t *testing.T) {
	check := health.SystemCheck()(context.Background())

	assert.Equal(t, "system", check.Name)
	assert.Equal(t, health.StatusHealthy, check.Status)
	assert.Equal(t, "System is healthy", check.Message)
	assert.NotZero(t, check.Duration)
	assert.NotNil(t, check.Details)
	assert.Contains(t, check.Details, "goroutines")
	assert.Contains(t, check.Details, "memory_alloc")
	assert.Contains(t, check.Details, "memory_heap")
}

func TestDatabaseCheck(t *testing.T) {
	// Mock database that always succeeds
	mockDB := &mockDatabase{shouldFail: false}
	check := health.DatabaseCheck(mockDB)(context.Background())

	assert.Equal(t, "database", check.Name)
	assert.Equal(t, health.StatusHealthy, check.Status)
	assert.Equal(t, "Database connection is healthy", check.Message)
	assert.NotZero(t, check.Duration)
}

func TestDatabaseCheck_Failure(t *testing.T) {
	// Mock database that always fails
	mockDB := &mockDatabase{shouldFail: true}
	check := health.DatabaseCheck(mockDB)(context.Background())

	assert.Equal(t, "database", check.Name)
	assert.Equal(t, health.StatusUnhealthy, check.Status)
	assert.Equal(t, assert.AnError.Error(), check.Message)
	assert.NotZero(t, check.Duration)
}

func TestKafkaCheck(t *testing.T) {
	// Mock producer that always succeeds
	mockProducer := &mockProducer{shouldFail: false}
	check := health.KafkaCheck(mockProducer)(context.Background())

	assert.Equal(t, "kafka", check.Name)
	assert.Equal(t, health.StatusHealthy, check.Status)
	assert.Equal(t, "Kafka connection is healthy", check.Message)
	assert.NotZero(t, check.Duration)
}

func TestKafkaCheck_Failure(t *testing.T) {
	// Mock producer that always fails
	mockProducer := &mockProducer{shouldFail: true}
	check := health.KafkaCheck(mockProducer)(context.Background())

	assert.Equal(t, "kafka", check.Name)
	assert.Equal(t, health.StatusUnhealthy, check.Status)
	assert.Equal(t, assert.AnError.Error(), check.Message)
	assert.NotZero(t, check.Duration)
}

// Mock implementations for testing
type mockDatabase struct {
	shouldFail bool
}

func (m *mockDatabase) Ping() error {
	if m.shouldFail {
		return assert.AnError
	}
	return nil
}

type mockProducer struct {
	shouldFail bool
}

func (m *mockProducer) Close() error {
	if m.shouldFail {
		return assert.AnError
	}
	return nil
}
