package genkit

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// InterruptManager handles generation interrupts
type InterruptManager struct {
	logger      *logrus.Logger
	interrupts  map[string]*InterruptInfo
	mu          sync.RWMutex
}

// InterruptInfo represents information about an interrupt
type InterruptInfo struct {
	RequestID   string    `json:"request_id"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
	Source      string    `json:"source"` // "user", "system", "timeout"
	Acknowledged bool     `json:"acknowledged"`
}

// NewInterruptManager creates a new interrupt manager
func NewInterruptManager(logger *logrus.Logger) (*InterruptManager, error) {
	im := &InterruptManager{
		logger:     logger,
		interrupts: make(map[string]*InterruptInfo),
	}

	// Start cleanup routine for old interrupts
	go im.cleanupRoutine()

	return im, nil
}

// Interrupt marks a request as interrupted
func (im *InterruptManager) Interrupt(requestID, reason string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	interrupt := &InterruptInfo{
		RequestID: requestID,
		Reason:    reason,
		Timestamp: time.Now(),
		Source:    "user",
		Acknowledged: false,
	}

	im.interrupts[requestID] = interrupt
	im.logger.Infof("Request interrupted: %s (Reason: %s)", requestID, reason)

	return nil
}

// IsInterrupted checks if a request is interrupted
func (im *InterruptManager) IsInterrupted(requestID string) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	interrupt, exists := im.interrupts[requestID]
	return exists && !interrupt.Acknowledged
}

// GetInterrupt retrieves interrupt information for a request
func (im *InterruptManager) GetInterrupt(requestID string) (*InterruptInfo, bool) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	interrupt, exists := im.interrupts[requestID]
	return interrupt, exists
}

// AcknowledgeInterrupt marks an interrupt as acknowledged
func (im *InterruptManager) AcknowledgeInterrupt(requestID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	interrupt, exists := im.interrupts[requestID]
	if !exists {
		return nil // Already cleaned up or never existed
	}

	interrupt.Acknowledged = true
	im.logger.Infof("Interrupt acknowledged: %s", requestID)

	return nil
}

// ListInterrupts returns all active interrupts
func (im *InterruptManager) ListInterrupts() []*InterruptInfo {
	im.mu.RLock()
	defer im.mu.RUnlock()

	interrupts := make([]*InterruptInfo, 0, len(im.interrupts))
	for _, interrupt := range im.interrupts {
		interrupts = append(interrupts, interrupt)
	}

	return interrupts
}

// ClearInterrupt removes an interrupt
func (im *InterruptManager) ClearInterrupt(requestID string) {
	im.mu.Lock()
	defer im.mu.Unlock()

	delete(im.interrupts, requestID)
	im.logger.Debugf("Cleared interrupt: %s", requestID)
}

// InterruptWithTimeout sets up an automatic timeout interrupt
func (im *InterruptManager) InterruptWithTimeout(requestID string, timeout time.Duration) {
	go func() {
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		<-timer.C

		// Check if request is still active
		if !im.IsInterrupted(requestID) {
			im.Interrupt(requestID, "timeout")
			im.logger.Warnf("Request timed out: %s (after %v)", requestID, timeout)
		}
	}()
}

// InterruptContext creates a context that gets cancelled when interrupted
func (im *InterruptManager) InterruptContext(ctx context.Context, requestID string) (context.Context, context.CancelFunc) {
	newCtx, cancel := context.WithCancel(ctx)

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if im.IsInterrupted(requestID) {
					cancel()
					return
				}
			case <-newCtx.Done():
				return
			}
		}
	}()

	return newCtx, cancel
}

// CreateInterruptHandler creates an interrupt handler for streaming operations
func (im *InterruptManager) CreateInterruptHandler(requestID string) func() bool {
	return func() bool {
		return im.IsInterrupted(requestID)
	}
}

// cleanupRoutine periodically cleans up old acknowledged interrupts
func (im *InterruptManager) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		im.cleanupOldInterrupts()
	}
}

// cleanupOldInterrupts removes old acknowledged interrupts
func (im *InterruptManager) cleanupOldInterrupts() {
	im.mu.Lock()
	defer im.mu.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	var toDelete []string

	for requestID, interrupt := range im.interrupts {
		if interrupt.Acknowledged && interrupt.Timestamp.Before(cutoff) {
			toDelete = append(toDelete, requestID)
		}
	}

	for _, requestID := range toDelete {
		delete(im.interrupts, requestID)
	}

	if len(toDelete) > 0 {
		im.logger.Debugf("Cleaned up %d old interrupts", len(toDelete))
	}
}

// InterruptAll interrupts all active requests
func (im *InterruptManager) InterruptAll(reason string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	count := 0
	for requestID, interrupt := range im.interrupts {
		if !interrupt.Acknowledged {
			interrupt.Reason = reason
			interrupt.Timestamp = time.Now()
			interrupt.Source = "system"
			count++
		}
	}

	im.logger.Infof("Interrupted %d active requests (Reason: %s)", count, reason)
	return nil
}

// GetStats returns interrupt statistics
func (im *InterruptManager) GetStats() map[string]interface{} {
	im.mu.RLock()
	defer im.mu.RUnlock()

	stats := map[string]interface{}{
		"total_interrupts": len(im.interrupts),
		"active_interrupts": 0,
		"acknowledged_interrupts": 0,
		"by_source": map[string]int{
			"user":    0,
			"system":  0,
			"timeout": 0,
		},
	}

	for _, interrupt := range im.interrupts {
		if interrupt.Acknowledged {
			stats["acknowledged_interrupts"] = stats["acknowledged_interrupts"].(int) + 1
		} else {
			stats["active_interrupts"] = stats["active_interrupts"].(int) + 1
		}

		sourceMap := stats["by_source"].(map[string]int)
		sourceMap[interrupt.Source]++
	}

	return stats
}