package lifecycle

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Manager orchestrates application lifecycle management
type Manager struct {
	components map[string]ManagedComponent
	statuses   map[string]*ComponentStatus
	mu         sync.RWMutex
	state      ComponentState
	shutdownCh chan struct{}
	done       chan struct{}
}

// NewManager creates a new lifecycle manager
func NewManager() *Manager {
	return &Manager{
		components: make(map[string]ManagedComponent),
		statuses:   make(map[string]*ComponentStatus),
		state:      StateUninitialized,
		shutdownCh: make(chan struct{}),
		done:       make(chan struct{}),
	}
}

// Register adds a component to the lifecycle manager
func (m *Manager) Register(component ManagedComponent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := component.Name()
	if _, exists := m.components[name]; exists {
		return fmt.Errorf("component %s already registered", name)
	}

	m.components[name] = component
	m.statuses[name] = &ComponentStatus{
		Name:  name,
		State: StateUninitialized,
	}

	log.WithField("component", name).Info("Component registered")
	return nil
}

// Initialize initializes all registered components
func (m *Manager) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state != StateUninitialized {
		return fmt.Errorf("manager already initialized")
	}

	m.state = StateInitializing
	log.Info("Initializing application components...")

	// Sort components by priority (lower numbers initialize first)
	components := m.getSortedComponents(false)

	for _, component := range components {
		name := component.Name()
		status := m.statuses[name]
		
		log.WithField("component", name).Info("Initializing component")
		status.State = StateInitializing
		status.StartTime = time.Now()

		if err := component.Initialize(ctx); err != nil {
			status.State = StateError
			status.LastError = err
			log.WithFields(log.Fields{
				"component": name,
				"error":     err,
			}).Error("Failed to initialize component")
			
			// Cleanup already initialized components
			m.shutdownInitializedComponents(ctx)
			return fmt.Errorf("failed to initialize component %s: %w", name, err)
		}

		status.State = StateReady
		log.WithField("component", name).Info("Component initialized successfully")
	}

	m.state = StateReady
	log.Info("All components initialized successfully")
	return nil
}

// Shutdown gracefully shuts down all components
func (m *Manager) Shutdown(ctx context.Context, options ShutdownOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state == StateShuttingDown || m.state == StateShutdown {
		return nil
	}

	m.state = StateShuttingDown
	log.Info("Shutting down application components...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	// Sort components by priority (higher numbers shutdown first)
	components := m.getSortedComponents(true)

	var shutdownErrors []error

	for _, component := range components {
		name := component.Name()
		status := m.statuses[name]
		
		if status.State != StateReady {
			continue
		}

		log.WithField("component", name).Info("Shutting down component")
		status.State = StateShuttingDown
		status.ShutdownTime = time.Now()

		if err := component.Shutdown(shutdownCtx); err != nil {
			status.State = StateError
			status.LastError = err
			shutdownErrors = append(shutdownErrors, fmt.Errorf("component %s: %w", name, err))
			log.WithFields(log.Fields{
				"component": name,
				"error":     err,
			}).Error("Failed to shutdown component")
		} else {
			status.State = StateShutdown
			log.WithField("component", name).Info("Component shutdown successfully")
		}
	}

	m.state = StateShutdown
	close(m.done)

	if len(shutdownErrors) > 0 {
		return fmt.Errorf("shutdown errors: %v", shutdownErrors)
	}

	log.Info("All components shutdown successfully")
	return nil
}

// HealthCheck runs health checks on all components
func (m *Manager) HealthCheck(ctx context.Context) map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make(map[string]error)

	for name, component := range m.components {
		status := m.statuses[name]
		
		if status.State != StateReady {
			results[name] = fmt.Errorf("component not ready: %s", status.State)
			continue
		}

		if err := component.HealthCheck(ctx); err != nil {
			results[name] = err
			status.LastError = err
		}
		
		status.LastHealthCheck = time.Now()
	}

	return results
}

// GetStatus returns the status of all components
func (m *Manager) GetStatus() map[string]ComponentStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make(map[string]ComponentStatus)
	for name, status := range m.statuses {
		statuses[name] = *status
	}

	return statuses
}

// GetComponentStatus returns the status of a specific component
func (m *Manager) GetComponentStatus(name string) (ComponentStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.statuses[name]
	if !exists {
		return ComponentStatus{}, false
	}

	return *status, true
}

// IsReady returns true if all components are ready
func (m *Manager) IsReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.state != StateReady {
		return false
	}

	for _, status := range m.statuses {
		if status.State != StateReady {
			return false
		}
	}

	return true
}

// Wait blocks until the manager is shutdown
func (m *Manager) Wait() {
	<-m.done
}

// getSortedComponents returns components sorted by priority
func (m *Manager) getSortedComponents(reverse bool) []ManagedComponent {
	components := make([]ManagedComponent, 0, len(m.components))
	for _, component := range m.components {
		components = append(components, component)
	}

	sort.Slice(components, func(i, j int) bool {
		if reverse {
			return components[i].Priority() > components[j].Priority()
		}
		return components[i].Priority() < components[j].Priority()
	})

	return components
}

// shutdownInitializedComponents shuts down components that were already initialized
func (m *Manager) shutdownInitializedComponents(ctx context.Context) {
	log.Info("Shutting down already initialized components due to initialization failure")
	
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Sort components by priority (higher numbers shutdown first)
	components := m.getSortedComponents(true)

	for _, component := range components {
		name := component.Name()
		status := m.statuses[name]
		
		if status.State != StateReady {
			continue
		}

		log.WithField("component", name).Info("Shutting down component")
		status.State = StateShuttingDown

		if err := component.Shutdown(shutdownCtx); err != nil {
			status.State = StateError
			status.LastError = err
			log.WithFields(log.Fields{
				"component": name,
				"error":     err,
			}).Error("Failed to shutdown component during cleanup")
		} else {
			status.State = StateShutdown
		}
	}
}

// StartHealthCheckMonitor starts a background health check monitor
func (m *Manager) StartHealthCheckMonitor(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-m.shutdownCh:
				return
			case <-ticker.C:
				results := m.HealthCheck(ctx)
				for name, err := range results {
					if err != nil {
						log.WithFields(log.Fields{
							"component": name,
							"error":     err,
						}).Warn("Component health check failed")
					}
				}
			}
		}
	}()
}

// StopHealthCheckMonitor stops the health check monitor
func (m *Manager) StopHealthCheckMonitor() {
	close(m.shutdownCh)
}