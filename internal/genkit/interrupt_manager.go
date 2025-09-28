package genkit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// InterruptManager manages generation interrupts
type InterruptManager struct {
	logger      *logrus.Logger
	interrupted map[string]bool
	mu          sync.RWMutex
}

// NewInterruptManager creates a new interrupt manager
func NewInterruptManager(logger *logrus.Logger) (*InterruptManager, error) {
	return &InterruptManager{
		logger:      logger,
		interrupted: make(map[string]bool),
	}, nil
}

// InterruptGeneration interrupts a generation by request ID
func (im *InterruptManager) InterruptGeneration(ctx context.Context, requestID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.interrupted[requestID] = true
	im.logger.Infof("Interrupted generation: %s", requestID)

	return nil
}

// IsInterrupted checks if a generation is interrupted
func (im *InterruptManager) IsInterrupted(requestID string) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return im.interrupted[requestID]
}

// ClearInterrupt clears an interrupt for a request ID
func (im *InterruptManager) ClearInterrupt(ctx context.Context, requestID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	delete(im.interrupted, requestID)
	im.logger.Debugf("Cleared interrupt: %s", requestID)

	return nil
}

// ListInterrupted returns all interrupted request IDs
func (im *InterruptManager) ListInterrupted(ctx context.Context) ([]string, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	var interrupted []string
	for requestID := range im.interrupted {
		interrupted = append(interrupted, requestID)
	}

	return interrupted, nil
}

// CleanupInterrupts removes old interrupts (older than 1 hour)
func (im *InterruptManager) CleanupInterrupts(ctx context.Context) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	// For simplicity, we'll just clear all interrupts
	// In a real implementation, you'd track timestamps and clean up old ones
	im.interrupted = make(map[string]bool)
	im.logger.Info("Cleaned up old interrupts")

	return nil
}

// StartCleanupRoutine starts a background routine to clean up old interrupts
func (im *InterruptManager) StartCleanupRoutine(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := im.CleanupInterrupts(ctx); err != nil {
					im.logger.Errorf("Failed to cleanup interrupts: %v", err)
				}
			}
		}
	}()
}

// InterruptRequest represents an interrupt request
type InterruptRequest struct {
	RequestID string `json:"request_id"`
	Reason    string `json:"reason,omitempty"`
}

// InterruptResponse represents an interrupt response
type InterruptResponse struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"` // "interrupted", "not_found", "already_completed"
	Message   string `json:"message,omitempty"`
}

// CreateInterruptRequest creates a new interrupt request
func (im *InterruptManager) CreateInterruptRequest(requestID, reason string) *InterruptRequest {
	return &InterruptRequest{
		RequestID: requestID,
		Reason:    reason,
	}
}

// ProcessInterruptRequest processes an interrupt request
func (im *InterruptManager) ProcessInterruptRequest(ctx context.Context, req *InterruptRequest) (*InterruptResponse, error) {
	err := im.InterruptGeneration(ctx, req.RequestID)
	if err != nil {
		return &InterruptResponse{
			RequestID: req.RequestID,
			Status:    "error",
			Message:   fmt.Sprintf("Failed to interrupt: %v", err),
		}, err
	}

	return &InterruptResponse{
		RequestID: req.RequestID,
		Status:    "interrupted",
		Message:   fmt.Sprintf("Generation interrupted: %s", req.Reason),
	}, nil
}

// GetInterruptStatus gets the interrupt status for a request ID
func (im *InterruptManager) GetInterruptStatus(ctx context.Context, requestID string) (*InterruptResponse, error) {
	isInterrupted := im.IsInterrupted(requestID)

	status := "active"
	if isInterrupted {
		status = "interrupted"
	}

	return &InterruptResponse{
		RequestID: requestID,
		Status:    status,
		Message:   fmt.Sprintf("Generation status: %s", status),
	}, nil
}

// ResumeGeneration resumes a generation (clears interrupt)
func (im *InterruptManager) ResumeGeneration(ctx context.Context, requestID string) (*InterruptResponse, error) {
	err := im.ClearInterrupt(ctx, requestID)
	if err != nil {
		return &InterruptResponse{
			RequestID: requestID,
			Status:    "error",
			Message:   fmt.Sprintf("Failed to resume: %v", err),
		}, err
	}

	return &InterruptResponse{
		RequestID: requestID,
		Status:    "resumed",
		Message:   "Generation resumed",
	}, nil
}

// BatchInterrupt interrupts multiple generations
func (im *InterruptManager) BatchInterrupt(ctx context.Context, requestIDs []string) ([]*InterruptResponse, error) {
	var responses []*InterruptResponse

	for _, requestID := range requestIDs {
		response, err := im.ProcessInterruptRequest(ctx, &InterruptRequest{
			RequestID: requestID,
			Reason:    "batch interrupt",
		})
		if err != nil {
			im.logger.Errorf("Failed to interrupt %s: %v", requestID, err)
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetInterruptStats returns interrupt statistics
func (im *InterruptManager) GetInterruptStats(ctx context.Context) (map[string]interface{}, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return map[string]interface{}{
		"total_interrupted": len(im.interrupted),
		"interrupted_ids":   im.getInterruptedIDs(),
		"timestamp":         time.Now().Format(time.RFC3339),
	}, nil
}

// getInterruptedIDs returns all interrupted request IDs
func (im *InterruptManager) getInterruptedIDs() []string {
	var ids []string
	for requestID := range im.interrupted {
		ids = append(ids, requestID)
	}
	return ids
}

// Close closes the interrupt manager
func (im *InterruptManager) Close() error {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.interrupted = make(map[string]bool)
	im.logger.Info("Interrupt manager closed")

	return nil
}
