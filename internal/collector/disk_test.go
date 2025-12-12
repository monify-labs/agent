package collector

import (
	"context"
	"testing"
	"time"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestDiskCollector_IOBandwidth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping disk collector test in short mode")
	}

	c := NewDiskCollector(true)
	defer c.Stop()

	ctx := context.Background()

	// First collection - establishes baseline
	result1, err := c.Collect(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, result1)

	resultMap1, ok := result1.(map[string]interface{})
	assert.True(t, ok, "Expected map[string]interface{}")

	diskMetrics1, ok := resultMap1["disk"].([]models.DiskMetrics)
	assert.True(t, ok, "Expected []models.DiskMetrics")

	// Should have at least one disk
	if len(diskMetrics1) > 0 {
		// First collection should have 0 I/O rates (no previous data)
		for _, metric := range diskMetrics1 {
			assert.Equal(t, 0.0, metric.ReadRate)
			assert.Equal(t, 0.0, metric.WriteRate)
			assert.Equal(t, 0.0, metric.ReadRateMBps)
			assert.Equal(t, 0.0, metric.WriteRateMBps)
			assert.Equal(t, 0.0, metric.ReadIOPS)
			assert.Equal(t, 0.0, metric.WriteIOPS)

			// Should have basic info
			assert.NotEmpty(t, metric.Device)
			assert.NotEmpty(t, metric.MountPoint)
			assert.Greater(t, metric.Total, uint64(0))
		}
	}

	// Wait a bit for some disk activity
	time.Sleep(2 * time.Second)

	// Second collection - should have I/O rates
	result2, err := c.Collect(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, result2)

	resultMap2, ok := result2.(map[string]interface{})
	assert.True(t, ok, "Expected map[string]interface{}")

	diskMetrics2, ok := resultMap2["disk"].([]models.DiskMetrics)
	assert.True(t, ok, "Expected []models.DiskMetrics")

	if len(diskMetrics2) > 0 {
		// Second collection should have I/O rates calculated
		for _, metric := range diskMetrics2 {
			// I/O rates should be calculated (may be 0 if no activity)
			assert.GreaterOrEqual(t, metric.ReadRate, 0.0)
			assert.GreaterOrEqual(t, metric.WriteRate, 0.0)

			// MBps conversion should be consistent
			expectedReadMBps := metric.ReadRate / 1_000_000
			expectedWriteMBps := metric.WriteRate / 1_000_000
			assert.InDelta(t, expectedReadMBps, metric.ReadRateMBps, 0.001)
			assert.InDelta(t, expectedWriteMBps, metric.WriteRateMBps, 0.001)

			// IOPS should be calculated
			assert.GreaterOrEqual(t, metric.ReadIOPS, 0.0)
			assert.GreaterOrEqual(t, metric.WriteIOPS, 0.0)
		}
	}
}

func TestDiskCollector_BasicMetrics(t *testing.T) {
	c := NewDiskCollector(true)
	defer c.Stop()

	ctx := context.Background()
	result, err := c.Collect(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Expected map[string]interface{}")

	diskMetrics, ok := resultMap["disk"].([]models.DiskMetrics)
	assert.True(t, ok, "Expected []models.DiskMetrics")

	// Should have at least one disk partition
	assert.Greater(t, len(diskMetrics), 0)

	for _, metric := range diskMetrics {
		// Should have all required fields
		assert.NotEmpty(t, metric.Device)
		assert.NotEmpty(t, metric.MountPoint)
		assert.NotEmpty(t, metric.FSType)

		// Space metrics should be valid
		assert.Greater(t, metric.Total, uint64(0))
		assert.GreaterOrEqual(t, metric.Used, uint64(0))
		assert.GreaterOrEqual(t, metric.Free, uint64(0))
		assert.GreaterOrEqual(t, metric.UsedPercent, 0.0)
		assert.LessOrEqual(t, metric.UsedPercent, 100.0)

		// Cumulative counters should be non-negative
		assert.GreaterOrEqual(t, metric.ReadBytes, uint64(0))
		assert.GreaterOrEqual(t, metric.WriteBytes, uint64(0))
		assert.GreaterOrEqual(t, metric.ReadCount, uint64(0))
		assert.GreaterOrEqual(t, metric.WriteCount, uint64(0))
	}
}

func TestDiskCollector_Disabled(t *testing.T) {
	c := NewDiskCollector(false)
	defer c.Stop()

	ctx := context.Background()
	metrics, err := c.Collect(ctx)

	assert.NoError(t, err)
	assert.Nil(t, metrics)
}

func TestDiskCollector_ConcurrentAccess(t *testing.T) {
	// Test that concurrent Collect calls don't cause data races
	c := NewDiskCollector(true)
	defer c.Stop()

	ctx := context.Background()

	// Run multiple collections concurrently
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			_, err := c.Collect(ctx)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestDiskCollector_MultipleCollections(t *testing.T) {
	// Test multiple sequential collections to verify state management
	c := NewDiskCollector(true)
	defer c.Stop()

	ctx := context.Background()

	// Perform 3 collections with delays
	for i := 0; i < 3; i++ {
		result, err := c.Collect(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok, "Expected map[string]interface{}")

		diskMetrics, ok := resultMap["disk"].([]models.DiskMetrics)
		assert.True(t, ok, "Expected []models.DiskMetrics")
		assert.Greater(t, len(diskMetrics), 0)

		if i > 0 {
			// After first collection, rates should be calculated
			for _, metric := range diskMetrics {
				// Rates should exist (even if 0)
				assert.GreaterOrEqual(t, metric.ReadRate, 0.0)
				assert.GreaterOrEqual(t, metric.WriteRate, 0.0)
			}
		}

		time.Sleep(500 * time.Millisecond)
	}
}
