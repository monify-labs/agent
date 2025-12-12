package collector

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestNetworkCollector_BandwidthCalculation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network collector test in short mode")
	}

	c := NewNetworkCollector(true)
	defer c.Stop()

	ctx := context.Background()

	// First collection - establishes baseline
	result1, err := c.Collect(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, result1)

	resultMap1, ok := result1.(map[string]interface{})
	assert.True(t, ok, "Expected map[string]interface{}")

	networkMetrics1, ok := resultMap1["network"].([]models.NetworkMetrics)
	assert.True(t, ok, "Expected []models.NetworkMetrics")

	// Should have at least one interface (even if it's just loopback excluded)
	if len(networkMetrics1) > 0 {
		// First collection should have 0 bandwidth rates (no previous data)
		for _, metric := range networkMetrics1 {
			assert.Equal(t, 0.0, metric.SendRate)
			assert.Equal(t, 0.0, metric.RecvRate)
			assert.Equal(t, 0.0, metric.SendRateMbps)
			assert.Equal(t, 0.0, metric.RecvRateMbps)
			assert.NotEmpty(t, metric.Interface)
			assert.Contains(t, []string{"public", "private"}, metric.Type)
		}
	}

	// Wait a bit for some network activity
	time.Sleep(2 * time.Second)

	// Second collection - should have bandwidth rates
	result2, err := c.Collect(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, result2)

	resultMap2, ok := result2.(map[string]interface{})
	assert.True(t, ok, "Expected map[string]interface{}")

	networkMetrics2, ok := resultMap2["network"].([]models.NetworkMetrics)
	assert.True(t, ok, "Expected []models.NetworkMetrics")

	if len(networkMetrics2) > 0 {
		// Second collection should have bandwidth rates calculated
		for _, metric := range networkMetrics2 {
			// Bandwidth rates should be calculated (may be 0 if no traffic)
			assert.GreaterOrEqual(t, metric.SendRate, 0.0)
			assert.GreaterOrEqual(t, metric.RecvRate, 0.0)

			// Mbps conversion should be consistent
			expectedSendMbps := (metric.SendRate * 8) / 1_000_000
			expectedRecvMbps := (metric.RecvRate * 8) / 1_000_000
			assert.InDelta(t, expectedSendMbps, metric.SendRateMbps, 0.001)
			assert.InDelta(t, expectedRecvMbps, metric.RecvRateMbps, 0.001)
		}
	}
}

func TestNetworkCollector_InterfaceClassification(t *testing.T) {
	c := NewNetworkCollector(true)
	defer c.Stop()

	ctx := context.Background()
	result, err := c.Collect(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Expected map[string]interface{}")

	networkMetrics, ok := resultMap["network"].([]models.NetworkMetrics)
	assert.True(t, ok, "Expected []models.NetworkMetrics")

	for _, metric := range networkMetrics {
		// Each interface should be classified as public or private
		assert.Contains(t, []string{"public", "private"}, metric.Type)

		// Should have all required fields
		assert.NotEmpty(t, metric.Interface)
		assert.GreaterOrEqual(t, metric.BytesSent, uint64(0))
		assert.GreaterOrEqual(t, metric.BytesRecv, uint64(0))
	}
}

func TestNetworkCollector_Disabled(t *testing.T) {
	c := NewNetworkCollector(false)
	defer c.Stop()

	ctx := context.Background()
	metrics, err := c.Collect(ctx)

	assert.NoError(t, err)
	assert.Nil(t, metrics)
}

func TestIsPublicIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		// Private IPv4
		{"Private 10.x", "10.0.0.1", false},
		{"Private 172.16.x", "172.16.0.1", false},
		{"Private 192.168.x", "192.168.1.1", false},
		{"Link-local", "169.254.1.1", false},
		{"Loopback", "127.0.0.1", false},

		// Public IPv4
		{"Public Google DNS", "8.8.8.8", true},
		{"Public Cloudflare", "1.1.1.1", true},
		{"Public AWS", "52.1.1.1", true},

		// Private IPv6
		{"IPv6 loopback", "::1", false},
		{"IPv6 link-local", "fe80::1", false},
		{"IPv6 ULA", "fc00::1", false},
		{"IPv6 ULA fd", "fd00::1", false},

		// Public IPv6
		{"IPv6 public", "2001:4860:4860::8888", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			assert.NotNil(t, ip)
			result := isPublicIP(ip)
			assert.Equal(t, tt.expected, result, "IP: %s", tt.ip)
		})
	}
}

func TestNetworkCollector_ConcurrentAccess(t *testing.T) {
	// Test that concurrent Collect calls don't cause data races
	c := NewNetworkCollector(true)
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
