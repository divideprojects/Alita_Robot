package lifecycle

import (
	"context"
	"time"
)

// Shutdownable represents a component that can be gracefully shutdown
type Shutdownable interface {
	Shutdown(ctx context.Context) error
}

// Initializable represents a component that can be initialized
type Initializable interface {
	Initialize(ctx context.Context) error
}

// HealthCheckable represents a component that can report its health status
type HealthCheckable interface {
	HealthCheck(ctx context.Context) error
}

// Component represents a managed component in the application lifecycle
type Component interface {
	// Name returns the unique name of the component
	Name() string
	
	// Priority returns the shutdown priority (higher numbers shutdown first)
	Priority() int
}

// ManagedComponent combines all lifecycle interfaces
type ManagedComponent interface {
	Component
	Initializable
	Shutdownable
	HealthCheckable
}

// ComponentState represents the current state of a component
type ComponentState int

const (
	StateUninitialized ComponentState = iota
	StateInitializing
	StateReady
	StateShuttingDown
	StateShutdown
	StateError
)

// String returns the string representation of the component state
func (s ComponentState) String() string {
	switch s {
	case StateUninitialized:
		return "uninitialized"
	case StateInitializing:
		return "initializing"
	case StateReady:
		return "ready"
	case StateShuttingDown:
		return "shutting_down"
	case StateShutdown:
		return "shutdown"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// ComponentStatus holds the status information of a component
type ComponentStatus struct {
	Name         string
	State        ComponentState
	LastError    error
	LastHealthCheck time.Time
	StartTime    time.Time
	ShutdownTime time.Time
}

// ShutdownOptions contains options for graceful shutdown
type ShutdownOptions struct {
	// Timeout is the maximum time to wait for shutdown
	Timeout time.Duration
	
	// Force indicates whether to force shutdown after timeout
	Force bool
}

// DefaultShutdownOptions returns default shutdown options
func DefaultShutdownOptions() ShutdownOptions {
	return ShutdownOptions{
		Timeout: 30 * time.Second,
		Force:   true,
	}
}