package cache

import (
	"context"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/divideprojects/Alita_Robot/alita/lifecycle"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	"github.com/eko/gocache/lib/v4/store"
	redis_store "github.com/eko/gocache/store/redis/v4"
	ristretto_store "github.com/eko/gocache/store/ristretto/v4"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

/*
CacheLifecycleManager manages the lifecycle of the cache system.

It implements the ManagedComponent interface and provides graceful
initialization, shutdown, and health checking for the cache system.
*/
type CacheLifecycleManager struct {
	name            string
	priority        int
	state           lifecycle.ComponentState
	stateMutex      sync.RWMutex
	redisClient     *redis.Client
	shutdownTimeout time.Duration
	healthCheckFreq time.Duration
	stopHealthCheck chan struct{}
	healthCheckOnce sync.Once
}

/*
NewCacheLifecycleManager creates a new CacheLifecycleManager instance.

Returns a configured lifecycle manager with default settings.
*/
func NewCacheLifecycleManager() *CacheLifecycleManager {
	return &CacheLifecycleManager{
		name:            "cache",
		priority:        50,
		state:           lifecycle.StateUninitialized,
		shutdownTimeout: 30 * time.Second,
		healthCheckFreq: 30 * time.Second,
		stopHealthCheck: make(chan struct{}),
	}
}

/*
Name returns the unique name of the cache component.
*/
func (c *CacheLifecycleManager) Name() string {
	return c.name
}

/*
Priority returns the shutdown priority of the cache component.
Higher numbers shutdown first.
*/
func (c *CacheLifecycleManager) Priority() int {
	return c.priority
}

/*
getState returns the current state of the component.
*/
func (c *CacheLifecycleManager) getState() lifecycle.ComponentState {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()
	return c.state
}

/*
setState updates the current state of the component.
*/
func (c *CacheLifecycleManager) setState(state lifecycle.ComponentState) {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()
	c.state = state
	log.WithFields(log.Fields{
		"component": c.name,
		"state":     state.String(),
	}).Debug("Cache component state changed")
}

/*
Initialize sets up the cache system with proper error handling and fallback mechanisms.

It attempts to initialize both Redis and Ristretto caches, falling back to
Ristretto-only mode if Redis is unavailable. Returns error only if both
cache systems fail to initialize.
*/
func (c *CacheLifecycleManager) Initialize(ctx context.Context) error {
	c.setState(lifecycle.StateInitializing)

	log.WithField("component", c.name).Info("Initializing cache system")

	// Initialize cache with fallback mechanisms
	err := c.initializeCacheWithFallback(ctx)
	if err != nil {
		c.setState(lifecycle.StateError)
		log.WithFields(log.Fields{
			"component": c.name,
			"error":     err,
		}).Error("Failed to initialize cache system")
		return err
	}

	// Start health check monitoring
	c.startHealthCheckMonitoring()

	c.setState(lifecycle.StateReady)
	log.WithField("component", c.name).Info("Cache system initialized successfully")
	return nil
}

/*
initializeCacheWithFallback initializes the cache system with fallback logic.

Attempts to initialize both Redis and Ristretto, falls back to Ristretto-only
if Redis fails, and returns error only if both fail.
*/
func (c *CacheLifecycleManager) initializeCacheWithFallback(ctx context.Context) error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Try to initialize Ristretto cache first
	ristrettoCache, err := initRistrettoCache()
	if err != nil {
		log.WithError(err).Error("Failed to initialize Ristretto cache")
		return err
	}

	// Try to initialize Redis cache
	redisClient, err := c.initRedisWithTimeout(ctx)
	if err != nil {
		log.WithError(err).Warn("Failed to initialize Redis cache, falling back to Ristretto only")
		// Use only Ristretto cache
		return c.setupRistrettoOnlyCache(ristrettoCache)
	}

	// Both caches available, use chain cache
	c.redisClient = redisClient
	return c.setupChainCache(ristrettoCache, redisClient)
}

/*
initRedisWithTimeout initializes Redis with a timeout context.
*/
func (c *CacheLifecycleManager) initRedisWithTimeout(ctx context.Context) (*redis.Client, error) {
	// Create a timeout context for Redis initialization
	redisCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	select {
	case <-redisCtx.Done():
		return nil, redisCtx.Err()
	default:
		return initRedisCache()
	}
}

/*
setupRistrettoOnlyCache sets up the cache system with Ristretto only.
*/
func (c *CacheLifecycleManager) setupRistrettoOnlyCache(ristrettoCache *ristretto.Cache) error {
	ristrettoStore := ristretto_store.NewRistretto(ristrettoCache)
	cacheManager := cache.New[any](ristrettoStore)
	Marshal = marshaler.New(cacheManager)
	isCacheEnabled = true

	log.Info("Cache system initialized with Ristretto only")
	return nil
}

/*
setupChainCache sets up the cache system with both Redis and Ristretto.
*/
func (c *CacheLifecycleManager) setupChainCache(ristrettoCache *ristretto.Cache, redisClient *redis.Client) error {
	redisStore := redis_store.NewRedis(redisClient, store.WithExpiration(10*time.Minute))
	ristrettoStore := ristretto_store.NewRistretto(ristrettoCache)
	cacheManager := cache.NewChain(cache.New[any](ristrettoStore), cache.New[any](redisStore))

	Manager = cacheManager
	Marshal = marshaler.New(cacheManager)
	isCacheEnabled = true

	log.Info("Cache system initialized with Redis and Ristretto")
	return nil
}

/*
Shutdown gracefully shuts down the cache system.

Closes Redis connections and stops health check monitoring.
*/
func (c *CacheLifecycleManager) Shutdown(ctx context.Context) error {
	c.setState(lifecycle.StateShuttingDown)

	log.WithField("component", c.name).Info("Shutting down cache system")

	// Stop health check monitoring
	c.stopHealthCheckMonitoring()

	// Create shutdown timeout context
	shutdownCtx, cancel := context.WithTimeout(ctx, c.shutdownTimeout)
	defer cancel()

	var shutdownErr error

	// Disable cache first
	cacheMutex.Lock()
	isCacheEnabled = false
	cacheMutex.Unlock()

	// Close Redis connection if available
	if c.redisClient != nil {
		if err := c.redisClient.Close(); err != nil {
			log.WithError(err).Error("Failed to close Redis connection")
			shutdownErr = err
		} else {
			log.Debug("Redis connection closed successfully")
		}
	}

	// Clear global variables
	Marshal = nil
	Manager = nil

	// Wait for shutdown timeout or completion
	select {
	case <-shutdownCtx.Done():
		if shutdownCtx.Err() == context.DeadlineExceeded {
			log.WithField("component", c.name).Warn("Cache shutdown timed out")
		}
	default:
		// Shutdown completed successfully
	}

	c.setState(lifecycle.StateShutdown)
	log.WithField("component", c.name).Info("Cache system shutdown completed")
	return shutdownErr
}

/*
HealthCheck verifies the cache system's health and connectivity.

Returns error if cache is unhealthy or unavailable.
*/
func (c *CacheLifecycleManager) HealthCheck(ctx context.Context) error {
	if c.getState() != lifecycle.StateReady {
		return ErrCacheNotEnabled
	}

	// Create health check timeout context
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Perform cache health check
	healthCheckErr := make(chan error, 1)
	go func() {
		healthCheckErr <- HealthCheckCache()
	}()

	select {
	case <-healthCtx.Done():
		return healthCtx.Err()
	case err := <-healthCheckErr:
		if err != nil {
			log.WithFields(log.Fields{
				"component": c.name,
				"error":     err,
			}).Error("Cache health check failed")
			return err
		}
		return nil
	}
}

/*
startHealthCheckMonitoring starts background health check monitoring.
*/
func (c *CacheLifecycleManager) startHealthCheckMonitoring() {
	c.healthCheckOnce.Do(func() {
		go c.healthCheckLoop()
	})
}

/*
stopHealthCheckMonitoring stops background health check monitoring.
*/
func (c *CacheLifecycleManager) stopHealthCheckMonitoring() {
	select {
	case c.stopHealthCheck <- struct{}{}:
	default:
		// Channel might be closed or full
	}
}

/*
healthCheckLoop runs periodic health checks in the background.
*/
func (c *CacheLifecycleManager) healthCheckLoop() {
	ticker := time.NewTicker(c.healthCheckFreq)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := c.HealthCheck(ctx); err != nil {
				log.WithFields(log.Fields{
					"component": c.name,
					"error":     err,
				}).Warn("Periodic cache health check failed")
			}
			cancel()
		case <-c.stopHealthCheck:
			log.WithField("component", c.name).Debug("Stopping cache health check monitoring")
			return
		}
	}
}

/*
GetCacheStats returns cache statistics and status information.
*/
func (c *CacheLifecycleManager) GetCacheStats() map[string]interface{} {
	stats := map[string]interface{}{
		"enabled":     IsCacheEnabled(),
		"state":       c.getState().String(),
		"has_redis":   c.redisClient != nil,
		"has_marshal": Marshal != nil,
		"has_manager": Manager != nil,
	}

	// Add Redis stats if available
	if c.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if info, err := c.redisClient.Info(ctx, "memory").Result(); err == nil {
			stats["redis_info"] = info
		}

		if poolStats := c.redisClient.PoolStats(); poolStats != nil {
			stats["redis_pool_stats"] = map[string]interface{}{
				"hits":        poolStats.Hits,
				"misses":      poolStats.Misses,
				"timeouts":    poolStats.Timeouts,
				"total_conns": poolStats.TotalConns,
				"idle_conns":  poolStats.IdleConns,
				"stale_conns": poolStats.StaleConns,
			}
		}
	}

	return stats
}

/*
GetCacheLifecycleManager returns a new cache lifecycle manager instance.

This function provides a convenient way to create and configure a cache
lifecycle manager for use with the application lifecycle management system.
*/
func GetCacheLifecycleManager() *CacheLifecycleManager {
	return NewCacheLifecycleManager()
}
