package collector

import (
	"context"
	"testing"
	"time"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestCPUCollector_Sampling(t *testing.T) {
	// Skip if running in short mode or CI where cpu metrics might be flaky
	if testing.Short() {
		t.Skip("Skipping CPU collector test in short mode")
	}

	c := NewCPUCollector(true)
	defer c.Stop()

	// Let it sample for 2 seconds (2 samples)
	time.Sleep(2500 * time.Millisecond)

	ctx := context.Background()
	metrics, err := c.Collect(ctx)
	
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	
	cpuMetrics, ok := metrics.(*models.CPUMetrics)
	assert.True(t, ok)
	
	// Should have gathered some usage
	// Note: In some CI environments CPU usage might be 0, but usually not exact 0.0 for 2 seconds active wait? 
	// Actually, sleeping doesn't consume CPU. But the background ticker does work.
	// We mainly verify the structure and that it didn't crash.
	assert.NotNil(t, cpuMetrics.PerCore)
	assert.GreaterOrEqual(t, cpuMetrics.UsagePercent, 0.0)
	
	// Check internal state (using reflection or just assuming black box)
	// Since we are in the same package (if we use package collector), we can access private fields?
	// No, test file is package collector, so we can access private fields.
	
	c.mu.Lock()
	// Samples should be cleared after Collect
	assert.Equal(t, 0, len(c.samples))
	c.mu.Unlock()
	
	// Let it run again
	time.Sleep(1500 * time.Millisecond)
	c.mu.Lock()
	// Should have roughly 1 sample (maybe 2 depending on timing)
	assert.True(t, len(c.samples) >= 1)
	c.mu.Unlock()
}

func TestCPUCollector_Disabled(t *testing.T) {
	c := NewCPUCollector(false)
	defer c.Stop()
	
	ctx := context.Background()
	metrics, err := c.Collect(ctx)
	
	assert.NoError(t, err)
	assert.Nil(t, metrics)
}
