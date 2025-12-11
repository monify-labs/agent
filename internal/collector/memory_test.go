package collector

import (
	"context"
	"testing"
	"time"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestMemoryCollector_Sampling(t *testing.T) {
	c := NewMemoryCollector(true)
	defer c.Stop()

	// Let it sample for a bit
	time.Sleep(2500 * time.Millisecond)

	ctx := context.Background()
	metrics, err := c.Collect(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	memMetrics, ok := metrics.(*models.MemoryMetrics)
	assert.True(t, ok)

	assert.Greater(t, memMetrics.Total, uint64(0))
	assert.Greater(t, memMetrics.Used, uint64(0))

	// Verify buffer is cleared
	c.mu.Lock()
	assert.Equal(t, 0, len(c.samples))
	c.mu.Unlock()
}
