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
	result, err := c.Collect(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Expected map[string]interface{}")

	memMetrics, ok := resultMap["memory"].(*models.MemoryMetrics)
	assert.True(t, ok, "Expected *models.MemoryMetrics")

	swapMetrics, ok := resultMap["swap"].(*models.SwapMetrics)
	assert.True(t, ok, "Expected *models.SwapMetrics")

	assert.Greater(t, memMetrics.Total, uint64(0))
	assert.Greater(t, memMetrics.Used, uint64(0))

	// Verify swap metrics
	assert.NotNil(t, swapMetrics)

	// Verify buffer is cleared
	c.mu.Lock()
	assert.Equal(t, 0, len(c.samples))
	c.mu.Unlock()
}
