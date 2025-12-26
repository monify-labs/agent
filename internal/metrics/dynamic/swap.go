package dynamic

import (
	"context"

	"github.com/monify-labs/agent/pkg/models"
	"github.com/shirou/gopsutil/v4/mem"
)

// CollectSwap gathers swap memory usage (no sampling needed)
func CollectSwap(ctx context.Context) (*models.SwapMetrics, error) {
	swap, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	return &models.SwapMetrics{
		Total:       swap.Total,
		Used:        swap.Used,
		UsedPercent: swap.UsedPercent,
	}, nil
}
