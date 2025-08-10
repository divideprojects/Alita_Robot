package safety

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/utils"
	log "github.com/sirupsen/logrus"
)

// WorkerSafetyManager manages worker pool safety and limits
type WorkerSafetyManager struct {
	maxGoroutines      int
	currentGoroutines  int64
	operationTimeout   time.Duration
	concurrencyLimiter chan struct{}
	monitoringEnabled  bool
	statsEnabled       bool
	mutex              sync.RWMutex
	shutdownCtx        context.Context
	shutdownCancel     context.CancelFunc
	activeOperations   map[string]*OperationTracker
	operationsMutex    sync.RWMutex
}

// OperationTracker tracks individual operations for safety
type OperationTracker struct {
	ID         string
	StartTime  time.Time
	WorkerType string
	Context    context.Context
	Cancel     context.CancelFunc
	IsActive   bool
}

// SafetyMetrics holds safety-related metrics
type SafetyMetrics struct {
	ActiveGoroutines     int
	MaxGoroutines        int
	ActiveOperations     int
	TimeoutOperations    int64
	PanicOperations      int64
	AverageOperationTime time.Duration
	MemoryUsageMB        float64
	LastHealthCheck      time.Time
}

var (
	globalSafetyManager *WorkerSafetyManager
	once                sync.Once
)

// GetGlobalSafetyManager returns the global safety manager singleton
func GetGlobalSafetyManager() *WorkerSafetyManager {
	once.Do(func() {
		globalSafetyManager = NewWorkerSafetyManager()
		globalSafetyManager.Start()
	})
	return globalSafetyManager
}

// NewWorkerSafetyManager creates a new worker safety manager
func NewWorkerSafetyManager() *WorkerSafetyManager {
	ctx, cancel := context.WithCancel(context.Background())

	maxGoroutines := config.MaxConcurrentOperations
	if maxGoroutines == 0 {
		maxGoroutines = 50 // Fallback default
	}

	timeout := time.Duration(config.OperationTimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second // Fallback default
	}

	return &WorkerSafetyManager{
		maxGoroutines:      maxGoroutines,
		operationTimeout:   timeout,
		concurrencyLimiter: make(chan struct{}, maxGoroutines),
		monitoringEnabled:  config.EnablePerformanceMonitoring,
		statsEnabled:       config.EnableBackgroundStats,
		shutdownCtx:        ctx,
		shutdownCancel:     cancel,
		activeOperations:   make(map[string]*OperationTracker),
	}
}

// Start begins the safety monitoring
func (sm *WorkerSafetyManager) Start() {
	if sm.monitoringEnabled {
		go sm.monitoringLoop()
	}

	log.WithFields(log.Fields{
		"max_goroutines":     sm.maxGoroutines,
		"operation_timeout":  sm.operationTimeout,
		"monitoring_enabled": sm.monitoringEnabled,
		"stats_enabled":      sm.statsEnabled,
	}).Info("Worker safety manager started")
}

// AcquireWorkerSlot safely acquires a worker slot with timeout
func (sm *WorkerSafetyManager) AcquireWorkerSlot(workerType string, operationID string) (*OperationTracker, error) {
	// Check if we're shutting down
	select {
	case <-sm.shutdownCtx.Done():
		return nil, utils.ErrSafetyManagerShutdown
	default:
	}

	// Try to acquire a slot with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case sm.concurrencyLimiter <- struct{}{}:
		// Successfully acquired slot
	case <-ctx.Done():
		return nil, fmt.Errorf("%w: %s", utils.ErrWorkerSlotTimeout, workerType)
	case <-sm.shutdownCtx.Done():
		return nil, utils.ErrShutdownDuringAcquisition
	}

	// Create operation tracker
	opCtx, opCancel := context.WithTimeout(sm.shutdownCtx, sm.operationTimeout)

	tracker := &OperationTracker{
		ID:         operationID,
		StartTime:  time.Now(),
		WorkerType: workerType,
		Context:    opCtx,
		Cancel:     opCancel,
		IsActive:   true,
	}

	// Register the operation
	sm.operationsMutex.Lock()
	sm.activeOperations[operationID] = tracker
	sm.operationsMutex.Unlock()

	// Update goroutine count
	sm.mutex.Lock()
	sm.currentGoroutines++
	sm.mutex.Unlock()

	if sm.statsEnabled {
		log.WithFields(log.Fields{
			"worker_type":       workerType,
			"operation_id":      operationID,
			"active_goroutines": sm.currentGoroutines,
		}).Debug("Worker slot acquired")
	}

	return tracker, nil
}

// ReleaseWorkerSlot safely releases a worker slot
func (sm *WorkerSafetyManager) ReleaseWorkerSlot(tracker *OperationTracker) {
	if tracker == nil {
		return
	}

	// Cancel the operation context
	tracker.Cancel()

	// Mark as inactive
	tracker.IsActive = false

	// Remove from active operations
	sm.operationsMutex.Lock()
	delete(sm.activeOperations, tracker.ID)
	sm.operationsMutex.Unlock()

	// Update goroutine count
	sm.mutex.Lock()
	sm.currentGoroutines--
	sm.mutex.Unlock()

	// Release the slot
	<-sm.concurrencyLimiter

	if sm.statsEnabled {
		duration := time.Since(tracker.StartTime)
		log.WithFields(log.Fields{
			"worker_type":     tracker.WorkerType,
			"operation_id":    tracker.ID,
			"duration_ms":     duration.Milliseconds(),
			"remaining_slots": len(sm.concurrencyLimiter),
		}).Debug("Worker slot released")
	}
}

// SafeExecute executes a function with safety measures
func (sm *WorkerSafetyManager) SafeExecute(workerType string, operationID string, fn func(ctx context.Context) error) error {
	tracker, err := sm.AcquireWorkerSlot(workerType, operationID)
	if err != nil {
		return err
	}
	defer sm.ReleaseWorkerSlot(tracker)

	// Execute with panic recovery
	var result error
	func() {
		defer func() {
			if r := recover(); r != nil {
				result = fmt.Errorf("panic in %s operation %s: %v", workerType, operationID, r)
				log.WithFields(log.Fields{
					"worker_type":  workerType,
					"operation_id": operationID,
					"panic":        r,
					"stack":        string(debug.Stack()),
				}).Error("Operation panicked")
			}
		}()

		result = fn(tracker.Context)
	}()

	return result
}

// monitoringLoop continuously monitors system health
func (sm *WorkerSafetyManager) monitoringLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.performHealthCheck()
		case <-sm.shutdownCtx.Done():
			return
		}
	}
}

// performHealthCheck checks system health and logs warnings
func (sm *WorkerSafetyManager) performHealthCheck() {
	metrics := sm.GetSafetyMetrics()

	// Check goroutine count
	if metrics.ActiveGoroutines > int(float64(metrics.MaxGoroutines)*0.8) {
		log.WithFields(log.Fields{
			"active_goroutines": metrics.ActiveGoroutines,
			"max_goroutines":    metrics.MaxGoroutines,
			"usage_percent":     float64(metrics.ActiveGoroutines) / float64(metrics.MaxGoroutines) * 100,
		}).Warn("High goroutine usage detected")
	}

	// Check for long-running operations
	sm.checkLongRunningOperations()

	// Check memory usage
	if metrics.MemoryUsageMB > 1000 {
		log.WithField("memory_mb", metrics.MemoryUsageMB).Warn("High memory usage detected")
	}

	// Log health summary
	if sm.statsEnabled {
		log.WithFields(log.Fields{
			"active_goroutines":  metrics.ActiveGoroutines,
			"max_goroutines":     metrics.MaxGoroutines,
			"active_operations":  metrics.ActiveOperations,
			"memory_usage_mb":    metrics.MemoryUsageMB,
			"avg_operation_time": metrics.AverageOperationTime.Milliseconds(),
		}).Debug("Safety manager health check")
	}
}

// checkLongRunningOperations identifies and warns about long-running operations
func (sm *WorkerSafetyManager) checkLongRunningOperations() {
	sm.operationsMutex.RLock()
	defer sm.operationsMutex.RUnlock()

	now := time.Now()
	longRunningThreshold := sm.operationTimeout * 2 // Double the normal timeout

	for id, tracker := range sm.activeOperations {
		if tracker.IsActive && now.Sub(tracker.StartTime) > longRunningThreshold {
			log.WithFields(log.Fields{
				"operation_id": id,
				"worker_type":  tracker.WorkerType,
				"duration":     now.Sub(tracker.StartTime),
				"threshold":    longRunningThreshold,
			}).Warn("Long-running operation detected")
		}
	}
}

// GetSafetyMetrics returns current safety metrics
func (sm *WorkerSafetyManager) GetSafetyMetrics() SafetyMetrics {
	sm.mutex.RLock()
	currentGoroutines := sm.currentGoroutines
	sm.mutex.RUnlock()

	sm.operationsMutex.RLock()
	activeOperations := len(sm.activeOperations)
	sm.operationsMutex.RUnlock()

	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SafetyMetrics{
		ActiveGoroutines: int(currentGoroutines),
		MaxGoroutines:    sm.maxGoroutines,
		ActiveOperations: activeOperations,
		MemoryUsageMB:    float64(m.Alloc) / 1024 / 1024,
		LastHealthCheck:  time.Now(),
	}
}

// ValidateWorkerCount validates and adjusts worker count based on system limits
func (sm *WorkerSafetyManager) ValidateWorkerCount(requested int, workerType string) int {
	maxAllowed := sm.maxGoroutines / 4 // Don't allow any single worker type to use more than 25% of slots

	if requested > maxAllowed {
		log.WithFields(log.Fields{
			"worker_type": workerType,
			"requested":   requested,
			"max_allowed": maxAllowed,
			"adjusted":    maxAllowed,
		}).Warn("Worker count adjusted due to safety limits")
		return maxAllowed
	}

	if requested <= 0 {
		defaultCount := 2 // Minimum default
		log.WithFields(log.Fields{
			"worker_type": workerType,
			"requested":   requested,
			"adjusted":    defaultCount,
		}).Warn("Invalid worker count, using default")
		return defaultCount
	}

	return requested
}

// Stop gracefully shuts down the safety manager
func (sm *WorkerSafetyManager) Stop() {
	log.Info("Stopping worker safety manager")

	// Cancel all active operations
	sm.operationsMutex.RLock()
	for _, tracker := range sm.activeOperations {
		tracker.Cancel()
	}
	sm.operationsMutex.RUnlock()

	// Cancel shutdown context
	sm.shutdownCancel()

	// Wait for operations to complete with timeout
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			log.Warn("Timeout waiting for operations to complete during shutdown")
			return
		case <-ticker.C:
			sm.operationsMutex.RLock()
			activeCount := len(sm.activeOperations)
			sm.operationsMutex.RUnlock()

			if activeCount == 0 {
				log.Info("All operations completed, safety manager stopped")
				return
			}
		}
	}
}
