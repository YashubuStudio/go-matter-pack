package readers

import (
	"context"
	"errors"
	"fmt"

	"github.com/cybergarage/go-matter/internal/matterctrl"
	"github.com/cybergarage/go-matter/internal/mattermodel"
	"github.com/cybergarage/go-matter/internal/metrics"
)

const (
	bridgedDeviceBasicInfoClusterID uint32 = 0x0039
	bridgedReachableAttrID          uint32 = 0x0011
	reachabilityReaderName                 = "reachability"
	metricReachableKey                     = "reachable"
)

// ReachabilityReader reads the reachability flag from bridged device basic information.
type ReachabilityReader struct{}

var _ metrics.MetricReader = (*ReachabilityReader)(nil)

// Name returns the reader identifier.
func (r *ReachabilityReader) Name() string {
	return reachabilityReaderName
}

// Read reads the reachable attribute.
func (r *ReachabilityReader) Read(ctx context.Context, ctrl matterctrl.Controller, nodeID uint64, dev mattermodel.BridgedDevice) (map[string]any, error) {
	raw, err := ctrl.ReadAttribute(ctx, nodeID, dev.Endpoint, bridgedDeviceBasicInfoClusterID, bridgedReachableAttrID)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, errors.New("reachability is unavailable")
	}
	reachable, err := parseReachable(raw)
	if err != nil {
		return nil, err
	}
	return map[string]any{metricReachableKey: reachable}, nil
}

func parseReachable(raw any) (bool, error) {
	switch value := raw.(type) {
	case bool:
		return value, nil
	case *bool:
		if value == nil {
			return false, errors.New("reachability value is nil")
		}
		return *value, nil
	case int:
		if value == 0 {
			return false, nil
		}
		if value == 1 {
			return true, nil
		}
		return false, fmt.Errorf("reachability out of range: %d", value)
	case uint8:
		if value == 0 {
			return false, nil
		}
		if value == 1 {
			return true, nil
		}
		return false, fmt.Errorf("reachability out of range: %d", value)
	default:
		return false, fmt.Errorf("unsupported reachability type %T", raw)
	}
}
