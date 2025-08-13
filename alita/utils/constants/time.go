package constants

import "time"

// Common time durations used throughout the application
const (
	// Cache durations
	AdminCacheTTL   = 30 * time.Minute
	DefaultCacheTTL = 5 * time.Minute
	ShortCacheTTL   = 1 * time.Minute
	LongCacheTTL    = 1 * time.Hour

	// Update intervals
	UserUpdateInterval    = 5 * time.Minute
	ChatUpdateInterval    = 5 * time.Minute
	ChannelUpdateInterval = 5 * time.Minute

	// Timeout durations
	DefaultTimeout = 10 * time.Second
	ShortTimeout   = 5 * time.Second
	LongTimeout    = 30 * time.Second
	CaptchaTimeout = 30 * time.Second

	// Retry and delay durations
	RetryDelay     = 2 * time.Second
	WebhookLatency = 10 * time.Millisecond

	// Activity monitoring
	MetricsStaleThreshold = 5 * time.Minute
)
