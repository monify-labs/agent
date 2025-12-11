package collector

import (
	"context"
)

// Collector is the interface that all metric collectors must implement
type Collector interface {
	// Collect collects metrics and returns them
	Collect(ctx context.Context) (interface{}, error)

	// Name returns the name of this collector
	Name() string

	// Enabled returns whether this collector is enabled
	Enabled() bool

	// Stop stops any background processes used by the collector
	Stop() error
}

// BaseCollector provides common functionality for all collectors
type BaseCollector struct {
	name    string
	enabled bool
}

// NewBaseCollector creates a new base collector
func NewBaseCollector(name string, enabled bool) *BaseCollector {
	return &BaseCollector{
		name:    name,
		enabled: enabled,
	}
}

// Name returns the collector name
func (b *BaseCollector) Name() string {
	return b.name
}

// Enabled returns whether the collector is enabled
func (b *BaseCollector) Enabled() bool {
	return b.enabled
}

// Stop stops the collector and cleans up resources
func (b *BaseCollector) Stop() error {
	// Default implementation does nothing
	return nil
}
