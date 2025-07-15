package db

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/divideprojects/Alita_Robot/alita/lifecycle"
)

// MongoLifecycleManager manages MongoDB client lifecycle
type MongoLifecycleManager struct {
	client *mongo.Client
	name   string
}

// NewMongoLifecycleManager creates a new MongoDB lifecycle manager
func NewMongoLifecycleManager() *MongoLifecycleManager {
	return &MongoLifecycleManager{
		client: mongoClient,
		name:   "mongodb",
	}
}

// Name returns the component name
func (m *MongoLifecycleManager) Name() string {
	return m.name
}

// Priority returns the shutdown priority (higher numbers shutdown first)
// MongoDB should have high priority (100) to ensure it shuts down after other components
func (*MongoLifecycleManager) Priority() int {
	return 100
}

// Initialize validates and sets up the MongoDB connection
func (m *MongoLifecycleManager) Initialize(ctx context.Context) error {
	log.WithField("component", m.name).Info("Initializing MongoDB connection")

	// Check if client exists
	if m.client == nil {
		return fmt.Errorf("MongoDB client is not initialized")
	}

	// Create a timeout context for initialization
	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Validate connection by pinging the database
	if err := m.client.Ping(initCtx, readpref.Primary()); err != nil {
		log.WithFields(log.Fields{
			"component": m.name,
			"error":     err,
		}).Error("Failed to ping MongoDB during initialization")
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Verify database operations by listing collections
	if err := m.validateDatabaseAccess(initCtx); err != nil {
		log.WithFields(log.Fields{
			"component": m.name,
			"error":     err,
		}).Error("Failed to validate database access during initialization")
		return fmt.Errorf("failed to validate database access: %w", err)
	}

	log.WithField("component", m.name).Info("MongoDB connection initialized successfully")
	return nil
}

// Shutdown gracefully disconnects the MongoDB client
func (m *MongoLifecycleManager) Shutdown(ctx context.Context) error {
	log.WithField("component", m.name).Info("Shutting down MongoDB connection")

	// Check if client exists
	if m.client == nil {
		log.WithField("component", m.name).Warn("MongoDB client is already nil, skipping shutdown")
		return nil
	}

	// Create a timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Disconnect the MongoDB client following best practices
	if err := m.client.Disconnect(shutdownCtx); err != nil {
		log.WithFields(log.Fields{
			"component": m.name,
			"error":     err,
		}).Error("Failed to disconnect MongoDB client")
		return fmt.Errorf("failed to disconnect MongoDB client: %w", err)
	}

	log.WithField("component", m.name).Info("MongoDB connection shut down successfully")
	return nil
}

// HealthCheck verifies MongoDB connection health
func (m *MongoLifecycleManager) HealthCheck(ctx context.Context) error {
	// Check if client exists
	if m.client == nil {
		return fmt.Errorf("MongoDB client is not initialized")
	}

	// Create a timeout context for health check
	healthCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Use client.Ping() to verify connection health
	if err := m.client.Ping(healthCtx, readpref.Primary()); err != nil {
		log.WithFields(log.Fields{
			"component": m.name,
			"error":     err,
		}).Debug("MongoDB health check failed")
		return fmt.Errorf("MongoDB ping failed: %w", err)
	}

	// Additional health check: verify we can access the database
	if err := m.validateDatabaseAccess(healthCtx); err != nil {
		log.WithFields(log.Fields{
			"component": m.name,
			"error":     err,
		}).Debug("MongoDB database access validation failed")
		return fmt.Errorf("MongoDB database access validation failed: %w", err)
	}

	return nil
}

// validateDatabaseAccess validates that we can access the database
func (*MongoLifecycleManager) validateDatabaseAccess(ctx context.Context) error {
	// Try to list collections to verify database access
	if adminSettingsColl == nil {
		return fmt.Errorf("database collections are not initialized")
	}

	// Perform a simple operation to verify database access
	// Use EstimatedDocumentCount as it's lightweight and doesn't require actual documents
	_, err := adminSettingsColl.EstimatedDocumentCount(ctx)
	if err != nil {
		return fmt.Errorf("failed to perform database validation query: %w", err)
	}

	return nil
}

// GetMongoLifecycleManager returns a configured MongoDB lifecycle manager
// This function should be called to register the MongoDB component with the lifecycle manager
func GetMongoLifecycleManager() lifecycle.ManagedComponent {
	return NewMongoLifecycleManager()
}
