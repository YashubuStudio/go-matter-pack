package readers

import (
	"context"
	"errors"
	"fmt"

	"github.com/YashubuStudio/go-matter-pack/internal/matterctrl"
	"github.com/YashubuStudio/go-matter-pack/internal/mattermodel"
	"github.com/YashubuStudio/go-matter-pack/internal/metrics"
)

const (
	powerSourceClusterID      uint32 = 0x002F
	batPercentRemainingAttrID uint32 = 0x000C
	powerSourceReaderName            = "power_source"
	metricBatteryPercentKey          = "battery_percent"
)

// PowerSourceReader reads battery information from the Power Source cluster.
type PowerSourceReader struct{}

var _ metrics.MetricReader = (*PowerSourceReader)(nil)

// Name returns the reader identifier.
func (r *PowerSourceReader) Name() string {
	return powerSourceReaderName
}

// Read reads the battery percentage remaining.
func (r *PowerSourceReader) Read(ctx context.Context, ctrl matterctrl.Controller, nodeID uint64, dev mattermodel.BridgedDevice) (map[string]any, error) {
	raw, err := ctrl.ReadAttribute(ctx, nodeID, dev.Endpoint, powerSourceClusterID, batPercentRemainingAttrID)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, errors.New("battery percent remaining is unavailable")
	}
	percent, err := parsePercentRemaining(raw)
	if err != nil {
		return nil, err
	}
	return map[string]any{metricBatteryPercentKey: percent}, nil
}

func parsePercentRemaining(raw any) (float64, error) {
	switch value := raw.(type) {
	case uint8:
		return float64(value) / 2.0, nil
	case uint16:
		return float64(value) / 2.0, nil
	case int:
		if value < 0 {
			return 0, fmt.Errorf("battery percent remaining out of range: %d", value)
		}
		return float64(value) / 2.0, nil
	case int32:
		if value < 0 {
			return 0, fmt.Errorf("battery percent remaining out of range: %d", value)
		}
		return float64(value) / 2.0, nil
	case int64:
		if value < 0 {
			return 0, fmt.Errorf("battery percent remaining out of range: %d", value)
		}
		return float64(value) / 2.0, nil
	case float32:
		if value < 0 {
			return 0, fmt.Errorf("battery percent remaining out of range: %f", value)
		}
		return float64(value), nil
	case float64:
		if value < 0 {
			return 0, fmt.Errorf("battery percent remaining out of range: %f", value)
		}
		return value, nil
	default:
		return 0, fmt.Errorf("unsupported battery percent remaining type %T", raw)
	}
}
