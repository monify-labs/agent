package sender

import (
	"context"

	"github.com/monify-labs/agent/pkg/models"
)

// Sender is the interface for sending metrics to the server
type Sender interface {
	// Send sends a metric payload to the server
	Send(ctx context.Context, payload *models.MetricPayload) (*models.ServerResponse, error)

	// Close closes the sender and releases resources
	Close() error
}
