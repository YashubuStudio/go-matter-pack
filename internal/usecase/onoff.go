package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/YashubuStudio/go-matter-pack/internal/matterctrl"
	"github.com/YashubuStudio/go-matter-pack/internal/store"
)

const (
	onOffClusterID   uint32 = 0x0006
	onOffAttributeID uint32 = 0x0000
	onOffCommandID   uint32 = 0x01
	offCommandID     uint32 = 0x00
	toggleCommandID  uint32 = 0x02
)

// OnOffService executes On/Off/Toggle commands and reads on/off state.
type OnOffService struct {
	ctrl  matterctrl.Controller
	store store.Store
}

// NewOnOffService returns a new OnOffService.
func NewOnOffService(ctrl matterctrl.Controller, store store.Store) *OnOffService {
	return &OnOffService{ctrl: ctrl, store: store}
}

// On sends the On command to the target device.
func (s *OnOffService) On(ctx context.Context, uniqueID string) error {
	return s.invoke(ctx, uniqueID, onOffCommandID)
}

// Off sends the Off command to the target device.
func (s *OnOffService) Off(ctx context.Context, uniqueID string) error {
	return s.invoke(ctx, uniqueID, offCommandID)
}

// Toggle sends the Toggle command to the target device.
func (s *OnOffService) Toggle(ctx context.Context, uniqueID string) error {
	return s.invoke(ctx, uniqueID, toggleCommandID)
}

// State reads the OnOff attribute from the target device.
func (s *OnOffService) State(ctx context.Context, uniqueID string) (bool, error) {
	record, registry, err := s.lookupDevice(ctx, uniqueID)
	if err != nil {
		return false, err
	}

	nodeID := record.NodeID
	if nodeID == 0 {
		nodeID = registry.HubNodeID
	}
	raw, err := s.ctrl.ReadAttribute(ctx, nodeID, record.Endpoint, onOffClusterID, onOffAttributeID)
	if err != nil {
		return false, err
	}
	return parseOnOffState(raw)
}

func (s *OnOffService) invoke(ctx context.Context, uniqueID string, cmdID uint32) error {
	record, registry, err := s.lookupDevice(ctx, uniqueID)
	if err != nil {
		return err
	}

	nodeID := record.NodeID
	if nodeID == 0 {
		nodeID = registry.HubNodeID
	}
	_, err = s.ctrl.InvokeCommand(ctx, nodeID, record.Endpoint, onOffClusterID, cmdID, nil)
	return err
}

func (s *OnOffService) lookupDevice(ctx context.Context, uniqueID string) (store.DeviceRecord, store.Registry, error) {
	if s == nil {
		return store.DeviceRecord{}, store.Registry{}, errors.New("onoff service is nil")
	}
	if s.ctrl == nil {
		return store.DeviceRecord{}, store.Registry{}, errors.New("controller is nil")
	}
	if s.store == nil {
		return store.DeviceRecord{}, store.Registry{}, errors.New("store is nil")
	}
	if uniqueID == "" {
		return store.DeviceRecord{}, store.Registry{}, errors.New("unique id is required")
	}

	var registry store.Registry
	if err := s.store.Load(ctx, &registry); err != nil {
		return store.DeviceRecord{}, store.Registry{}, err
	}

	record, ok := registry.Find(uniqueID)
	if !ok {
		return store.DeviceRecord{}, store.Registry{}, fmt.Errorf("device not found for unique_id %s", uniqueID)
	}
	if record.Missing {
		return store.DeviceRecord{}, store.Registry{}, fmt.Errorf("device %s is marked missing", uniqueID)
	}
	if record.Endpoint == 0 {
		return store.DeviceRecord{}, store.Registry{}, errors.New("device endpoint is not set")
	}

	nodeID := record.NodeID
	if nodeID == 0 {
		nodeID = registry.HubNodeID
	}
	if nodeID == 0 {
		return store.DeviceRecord{}, store.Registry{}, errors.New("hub node id is not set")
	}
	return record, registry, nil
}

func parseOnOffState(raw any) (bool, error) {
	switch value := raw.(type) {
	case bool:
		return value, nil
	case *bool:
		if value == nil {
			return false, errors.New("onoff value is nil")
		}
		return *value, nil
	case int:
		if value == 0 {
			return false, nil
		}
		if value == 1 {
			return true, nil
		}
		return false, fmt.Errorf("onoff out of range: %d", value)
	case uint8:
		if value == 0 {
			return false, nil
		}
		if value == 1 {
			return true, nil
		}
		return false, fmt.Errorf("onoff out of range: %d", value)
	default:
		return false, fmt.Errorf("unsupported onoff type %T", raw)
	}
}
