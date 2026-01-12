package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/cybergarage/go-matter-pack/internal/matterctrl"
	"github.com/cybergarage/go-matter-pack/internal/mattermodel"
	"github.com/cybergarage/go-matter-pack/internal/metrics"
	"github.com/cybergarage/go-matter-pack/internal/store"
)

// OutputFormat represents the serialized output format for poll results.
type OutputFormat string

const (
	OutputFormatJSON  OutputFormat = "json"
	OutputFormatJSONL OutputFormat = "jsonl"
)

// PollService reads device metrics from the registry.
type PollService struct {
	ctrl    matterctrl.Controller
	store   store.Store
	readers []metrics.MetricReader
}

// PollResult represents a snapshot of metrics for all devices.
type PollResult struct {
	Timestamp time.Time           `json:"timestamp"`
	Devices   []PollDeviceMetrics `json:"devices"`
}

// PollDeviceMetrics represents metrics for a single device.
type PollDeviceMetrics struct {
	UniqueID string            `json:"unique_id"`
	Label    string            `json:"label,omitempty"`
	NodeID   uint64            `json:"node_id"`
	Endpoint uint16            `json:"endpoint"`
	Missing  bool              `json:"missing,omitempty"`
	Metrics  map[string]any    `json:"metrics,omitempty"`
	Errors   map[string]string `json:"errors,omitempty"`
}

// NewPollService returns a new PollService.
func NewPollService(ctrl matterctrl.Controller, store store.Store, readers ...metrics.MetricReader) *PollService {
	return &PollService{ctrl: ctrl, store: store, readers: readers}
}

// PollOnce reads metrics once and returns the result.
func (s *PollService) PollOnce(ctx context.Context) (PollResult, error) {
	if s == nil {
		return PollResult{}, errors.New("poll service is nil")
	}
	if s.ctrl == nil {
		return PollResult{}, errors.New("controller is nil")
	}
	if s.store == nil {
		return PollResult{}, errors.New("store is nil")
	}

	var registry store.Registry
	if err := s.store.Load(ctx, &registry); err != nil {
		return PollResult{}, err
	}

	result := PollResult{
		Timestamp: time.Now().UTC(),
		Devices:   []PollDeviceMetrics{},
	}

	if len(registry.Devices) == 0 {
		return result, nil
	}

	keys := make([]string, 0, len(registry.Devices))
	for key := range registry.Devices {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, uniqueID := range keys {
		record := registry.Devices[uniqueID]
		deviceMetrics := PollDeviceMetrics{
			UniqueID: record.UniqueID,
			Label:    record.Label,
			NodeID:   record.NodeID,
			Endpoint: record.Endpoint,
			Missing:  record.Missing,
		}

		if deviceMetrics.NodeID == 0 {
			deviceMetrics.NodeID = registry.HubNodeID
		}

		if record.Missing {
			deviceMetrics.Errors = map[string]string{"device": "device is marked missing"}
			result.Devices = append(result.Devices, deviceMetrics)
			continue
		}
		if deviceMetrics.NodeID == 0 {
			deviceMetrics.Errors = map[string]string{"device": "hub node id is not set"}
			result.Devices = append(result.Devices, deviceMetrics)
			continue
		}
		if deviceMetrics.Endpoint == 0 {
			deviceMetrics.Errors = map[string]string{"device": "device endpoint is not set"}
			result.Devices = append(result.Devices, deviceMetrics)
			continue
		}

		dev := mattermodel.BridgedDevice{
			NodeID:    deviceMetrics.NodeID,
			Endpoint:  deviceMetrics.Endpoint,
			UniqueID:  record.UniqueID,
			NodeLabel: record.NodeLabel,
			Reachable: record.Reachable,
		}

		metricsMap := make(map[string]any)
		errorsMap := make(map[string]string)
		for _, reader := range s.readers {
			if reader == nil {
				continue
			}
			values, err := reader.Read(ctx, s.ctrl, deviceMetrics.NodeID, dev)
			if err != nil {
				errorsMap[reader.Name()] = err.Error()
				continue
			}
			for key, value := range values {
				metricsMap[key] = value
			}
		}
		if len(metricsMap) > 0 {
			deviceMetrics.Metrics = metricsMap
		}
		if len(errorsMap) > 0 {
			deviceMetrics.Errors = errorsMap
		}
		result.Devices = append(result.Devices, deviceMetrics)
	}

	return result, nil
}

// Watch polls at the given interval until the context is canceled.
func (s *PollService) Watch(ctx context.Context, interval time.Duration, w io.Writer, format OutputFormat) error {
	if interval <= 0 {
		return fmt.Errorf("invalid interval: %s", interval)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		result, err := s.PollOnce(ctx)
		if err != nil {
			return err
		}
		if err := writePollResult(w, format, result); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// writePollResult writes the poll result to the writer in the requested format.
func writePollResult(w io.Writer, format OutputFormat, result PollResult) error {
	payload, err := json.Marshal(result)
	if err != nil {
		return err
	}
	if format == OutputFormatJSONL {
		payload = append(payload, '\n')
	} else if format != OutputFormatJSON {
		return fmt.Errorf("unsupported output format: %s", format)
	}
	_, err = w.Write(payload)
	return err
}
