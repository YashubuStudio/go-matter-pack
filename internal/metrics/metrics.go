package metrics

import (
	"context"

	"github.com/cybergarage/go-matter-pack/internal/matterctrl"
	"github.com/cybergarage/go-matter-pack/internal/mattermodel"
)

// MetricReader describes a plugin that reads metrics from a bridged device.
type MetricReader interface {
	Name() string
	Read(ctx context.Context, ctrl matterctrl.Controller, nodeID uint64, dev mattermodel.BridgedDevice) (map[string]any, error)
}
