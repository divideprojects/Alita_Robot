package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

// SignalHandler handles OS signals for graceful shutdown
type SignalHandler struct {
	manager    *Manager
	options    ShutdownOptions
	signalCh   chan os.Signal
	shutdownCh chan struct{}
}

// NewSignalHandler creates a new signal handler
func NewSignalHandler(manager *Manager, options ShutdownOptions) *SignalHandler {
	return &SignalHandler{
		manager:    manager,
		options:    options,
		signalCh:   make(chan os.Signal, 1),
		shutdownCh: make(chan struct{}),
	}
}

// Start starts listening for OS signals
func (h *SignalHandler) Start(ctx context.Context) {
	// Register signal handlers
	signal.Notify(h.signalCh, syscall.SIGINT, syscall.SIGTERM)

	go h.handleSignals(ctx)
	log.Info("Signal handler started")
}

// Stop stops the signal handler
func (h *SignalHandler) Stop() {
	signal.Stop(h.signalCh)
	close(h.shutdownCh)
	log.Info("Signal handler stopped")
}

// handleSignals processes incoming signals
func (h *SignalHandler) handleSignals(ctx context.Context) {
	for {
		select {
		case sig := <-h.signalCh:
			log.WithField("signal", sig.String()).Info("Received shutdown signal")
			h.initiateShutdown(ctx, sig)
			return
		case <-h.shutdownCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// initiateShutdown starts the graceful shutdown process
func (h *SignalHandler) initiateShutdown(ctx context.Context, sig os.Signal) {
	log.WithField("signal", sig.String()).Info("Initiating graceful shutdown")

	// Create shutdown context
	shutdownCtx, cancel := context.WithTimeout(ctx, h.options.Timeout)
	defer cancel()

	// Perform graceful shutdown
	if err := h.manager.Shutdown(shutdownCtx, h.options); err != nil {
		log.WithError(err).Error("Error during graceful shutdown")

		if h.options.Force {
			log.Error("Force shutdown initiated")
			h.forceShutdown()
		}
	} else {
		log.Info("Graceful shutdown completed successfully")
	}
}

// forceShutdown performs emergency shutdown
func (h *SignalHandler) forceShutdown() {
	log.Error("Performing emergency shutdown")

	// Give a small grace period for cleanup
	time.Sleep(1 * time.Second)

	// Force exit
	os.Exit(1)
}

// WaitForShutdown blocks until a shutdown signal is received
func (h *SignalHandler) WaitForShutdown(ctx context.Context) {
	select {
	case <-h.signalCh:
		// Signal received, shutdown will be handled by handleSignals
	case <-ctx.Done():
		// Context cancelled
	}
}

// GracefulShutdown provides a convenient way to setup and wait for graceful shutdown
func GracefulShutdown(ctx context.Context, manager *Manager, options ShutdownOptions) error {
	handler := NewSignalHandler(manager, options)
	handler.Start(ctx)
	defer handler.Stop()

	// Wait for shutdown signal
	handler.WaitForShutdown(ctx)

	// Wait for manager to finish shutdown
	manager.Wait()

	return nil
}

// SetupGracefulShutdown sets up graceful shutdown with default options
func SetupGracefulShutdown(ctx context.Context, manager *Manager) error {
	options := DefaultShutdownOptions()
	return GracefulShutdown(ctx, manager, options)
}
