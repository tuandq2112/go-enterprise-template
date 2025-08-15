package database

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Connection represents a database connection
type Connection interface {
	// Basic operations
	Ping(ctx context.Context) error
	Close() error
	IsValid() bool

	// Connection info
	GetID() string
	GetCreatedAt() time.Time
	GetLastUsed() time.Time
	GetUseCount() int64
}

// ConnectionFactory creates new database connections
type ConnectionFactory interface {
	CreateConnection(ctx context.Context) (Connection, error)
	ValidateConnection(ctx context.Context, conn Connection) error
}

// ConnectionPool manages a pool of database connections
type ConnectionPool struct {
	factory     ConnectionFactory
	connections chan Connection
	mu          sync.RWMutex
	config      *PoolConfig
	stats       *PoolStats
	closed      bool
}

// PoolConfig holds connection pool configuration
type PoolConfig struct {
	MaxOpenConns        int           // Maximum number of open connections
	MaxIdleConns        int           // Maximum number of idle connections
	ConnMaxLifetime     time.Duration // Maximum lifetime of connections
	ConnMaxIdleTime     time.Duration // Maximum idle time of connections
	ConnTimeout         time.Duration // Connection timeout
	HealthCheckInterval time.Duration // Health check interval
}

// DefaultPoolConfig returns default connection pool configuration
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxOpenConns:        25,
		MaxIdleConns:        5,
		ConnMaxLifetime:     5 * time.Minute,
		ConnMaxIdleTime:     5 * time.Minute,
		ConnTimeout:         30 * time.Second,
		HealthCheckInterval: 1 * time.Minute,
	}
}

// PoolStats holds connection pool statistics
type PoolStats struct {
	mu                 sync.RWMutex
	MaxOpenConnections int
	OpenConnections    int
	InUse              int
	Idle               int
	WaitCount          int64
	WaitDuration       time.Duration
	MaxIdleClosed      int64
	MaxLifetimeClosed  int64
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(factory ConnectionFactory, config *PoolConfig) *ConnectionPool {
	if config == nil {
		config = DefaultPoolConfig()
	}

	pool := &ConnectionPool{
		factory:     factory,
		connections: make(chan Connection, config.MaxOpenConns),
		config:      config,
		stats:       &PoolStats{MaxOpenConnections: config.MaxOpenConns},
	}

	// Start health checker
	go pool.healthChecker()

	return pool
}

// GetConnection gets a connection from the pool
func (cp *ConnectionPool) GetConnection(ctx context.Context) (Connection, error) {
	if cp.closed {
		return nil, fmt.Errorf("connection pool is closed")
	}

	// Try to get an existing connection
	select {
	case conn := <-cp.connections:
		if cp.isConnectionValid(conn) {
			cp.updateStats(conn, true)
			return conn, nil
		}
		// Invalid connection, close it and create a new one
		conn.Close()
		cp.decrementOpenConnections()
	default:
		// No available connections
	}

	// Check if we can create a new connection
	if cp.getOpenConnections() >= cp.config.MaxOpenConns {
		// Wait for a connection to become available
		cp.incrementWaitCount()
		start := time.Now()

		select {
		case conn := <-cp.connections:
			cp.updateWaitDuration(time.Since(start))
			if cp.isConnectionValid(conn) {
				cp.updateStats(conn, true)
				return conn, nil
			}
			conn.Close()
			cp.decrementOpenConnections()
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Create a new connection
	return cp.createConnection(ctx)
}

// ReturnConnection returns a connection to the pool
func (cp *ConnectionPool) ReturnConnection(conn Connection) {
	if cp.closed || conn == nil {
		if conn != nil {
			conn.Close()
		}
		return
	}

	// Update connection stats
	cp.updateStats(conn, false)

	// Check if connection is still valid
	if !cp.isConnectionValid(conn) {
		conn.Close()
		cp.decrementOpenConnections()
		return
	}

	// Check if connection has exceeded max lifetime
	if time.Since(conn.GetCreatedAt()) > cp.config.ConnMaxLifetime {
		conn.Close()
		cp.decrementOpenConnections()
		cp.incrementMaxLifetimeClosed()
		return
	}

	// Check if connection has exceeded max idle time
	if time.Since(conn.GetLastUsed()) > cp.config.ConnMaxIdleTime {
		conn.Close()
		cp.decrementOpenConnections()
		cp.incrementMaxIdleClosed()
		return
	}

	// Return connection to pool
	select {
	case cp.connections <- conn:
		// Successfully returned to pool
	default:
		// Pool is full, close the connection
		conn.Close()
		cp.decrementOpenConnections()
	}
}

// Close closes the connection pool
func (cp *ConnectionPool) Close() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if cp.closed {
		return nil
	}

	cp.closed = true

	// Close all connections in the pool
	close(cp.connections)
	for conn := range cp.connections {
		conn.Close()
	}

	return nil
}

// Stats returns connection pool statistics
func (cp *ConnectionPool) Stats() *PoolStats {
	cp.stats.mu.RLock()
	defer cp.stats.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := &PoolStats{
		MaxOpenConnections: cp.stats.MaxOpenConnections,
		OpenConnections:    cp.stats.OpenConnections,
		InUse:              cp.stats.InUse,
		Idle:               cp.stats.Idle,
		WaitCount:          cp.stats.WaitCount,
		WaitDuration:       cp.stats.WaitDuration,
		MaxIdleClosed:      cp.stats.MaxIdleClosed,
		MaxLifetimeClosed:  cp.stats.MaxLifetimeClosed,
	}

	return stats
}

// createConnection creates a new connection
func (cp *ConnectionPool) createConnection(ctx context.Context) (Connection, error) {
	conn, err := cp.factory.CreateConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	cp.incrementOpenConnections()
	return conn, nil
}

// isConnectionValid checks if a connection is valid
func (cp *ConnectionPool) isConnectionValid(conn Connection) bool {
	if conn == nil {
		return false
	}

	// Check if connection is valid
	if !conn.IsValid() {
		return false
	}

	// Validate connection with factory
	ctx, cancel := context.WithTimeout(context.Background(), cp.config.ConnTimeout)
	defer cancel()

	if err := cp.factory.ValidateConnection(ctx, conn); err != nil {
		return false
	}

	return true
}

// healthChecker periodically checks connection health
func (cp *ConnectionPool) healthChecker() {
	ticker := time.NewTicker(cp.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if cp.closed {
				return
			}
			cp.healthCheck()
		}
	}
}

// healthCheck performs health check on idle connections
func (cp *ConnectionPool) healthCheck() {
	// This is a simplified health check
	// In a real implementation, you might want to check idle connections
	// and remove invalid ones from the pool
}

// updateStats updates connection pool statistics
func (cp *ConnectionPool) updateStats(conn Connection, inUse bool) {
	cp.stats.mu.Lock()
	defer cp.stats.mu.Unlock()

	if inUse {
		cp.stats.InUse++
		cp.stats.Idle--
	} else {
		cp.stats.InUse--
		cp.stats.Idle++
	}
}

// getOpenConnections returns the number of open connections
func (cp *ConnectionPool) getOpenConnections() int {
	cp.stats.mu.RLock()
	defer cp.stats.mu.RUnlock()
	return cp.stats.OpenConnections
}

// incrementOpenConnections increments the open connections count
func (cp *ConnectionPool) incrementOpenConnections() {
	cp.stats.mu.Lock()
	defer cp.stats.mu.Unlock()
	cp.stats.OpenConnections++
}

// decrementOpenConnections decrements the open connections count
func (cp *ConnectionPool) decrementOpenConnections() {
	cp.stats.mu.Lock()
	defer cp.stats.mu.Unlock()
	if cp.stats.OpenConnections > 0 {
		cp.stats.OpenConnections--
	}
}

// incrementWaitCount increments the wait count
func (cp *ConnectionPool) incrementWaitCount() {
	cp.stats.mu.Lock()
	defer cp.stats.mu.Unlock()
	cp.stats.WaitCount++
}

// updateWaitDuration updates the wait duration
func (cp *ConnectionPool) updateWaitDuration(duration time.Duration) {
	cp.stats.mu.Lock()
	defer cp.stats.mu.Unlock()
	cp.stats.WaitDuration += duration
}

// incrementMaxIdleClosed increments the max idle closed count
func (cp *ConnectionPool) incrementMaxIdleClosed() {
	cp.stats.mu.Lock()
	defer cp.stats.mu.Unlock()
	cp.stats.MaxIdleClosed++
}

// incrementMaxLifetimeClosed increments the max lifetime closed count
func (cp *ConnectionPool) incrementMaxLifetimeClosed() {
	cp.stats.mu.Lock()
	defer cp.stats.mu.Unlock()
	cp.stats.MaxLifetimeClosed++
}
