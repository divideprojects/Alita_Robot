package context

import (
	"context"
	"time"

	"github.com/divideprojects/Alita_Robot/alita/config"
)

// WithTimeout creates a context with the default operation timeout
func WithTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	timeout := time.Duration(config.OperationTimeoutSeconds) * time.Second
	return context.WithTimeout(parent, timeout)
}

// WithConfigTimeout creates a context with the configured timeout
func WithConfigTimeout(parent context.Context, cfg *config.Config) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, cfg.OperationTimeout)
}

// WithShortTimeout creates a context with a short timeout for quick operations
func WithShortTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 5*time.Second)
}

// WithLongTimeout creates a context with a longer timeout for complex operations
func WithLongTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 2*time.Minute)
}

// Background returns a background context with the default timeout
func Background() (context.Context, context.CancelFunc) {
	return WithTimeout(context.Background())
}

// TODO returns a context.TODO with the default timeout
func TODO() (context.Context, context.CancelFunc) {
	return WithTimeout(context.TODO())
}
