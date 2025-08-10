package utils

import "errors"

// Worker safety errors
var (
	ErrSafetyManagerShutdown     = errors.New("SAFETY_MANAGER_SHUTDOWN")
	ErrWorkerSlotTimeout         = errors.New("WORKER_SLOT_TIMEOUT")
	ErrShutdownDuringAcquisition = errors.New("SHUTDOWN_DURING_ACQUISITION")
)

// Pipeline errors
var (
	ErrNoProcessorRegistered = errors.New("NO_PROCESSOR_REGISTERED")
	ErrPipelineShuttingDown  = errors.New("PIPELINE_SHUTTING_DOWN")
	ErrProcessorTimeout      = errors.New("PROCESSOR_TIMEOUT")
	ErrResultsTimeout        = errors.New("RESULTS_TIMEOUT")
)
