package mattermodel

import (
	"context"
	"errors"
	"fmt"
)

const (
	descriptorClusterID       uint32 = 0x001D
	descriptorPartsListAttrID uint32 = 0x0003

	bridgedDeviceBasicInfoClusterID uint32 = 0x0039
	bridgedNodeLabelAttrID          uint32 = 0x0005
	bridgedReachableAttrID          uint32 = 0x0011
	bridgedUniqueIDAttrID           uint32 = 0x0012
)

// Controller defines the minimal interface required to scan bridged devices.
type Controller interface {
	ReadAttribute(ctx context.Context, nodeID uint64, endpoint uint16, clusterID uint32, attrID uint32) (any, error)
}

// ScanBridgedDevices enumerates endpoints from PartsList and reads bridged device attributes.
func ScanBridgedDevices(ctx context.Context, ctrl Controller, nodeID uint64) ([]BridgedDevice, error) {
	parts, err := readPartsList(ctx, ctrl, nodeID)
	if err != nil {
		return nil, err
	}

	devices := make([]BridgedDevice, 0, len(parts))
	for _, endpoint := range parts {
		device, ok, err := readBridgedDevice(ctx, ctrl, nodeID, endpoint)
		if err != nil {
			return nil, err
		}
		if ok {
			devices = append(devices, device)
		}
	}
	return devices, nil
}

func readPartsList(ctx context.Context, ctrl Controller, nodeID uint64) ([]uint16, error) {
	raw, err := ctrl.ReadAttribute(ctx, nodeID, 0, descriptorClusterID, descriptorPartsListAttrID)
	if err != nil {
		return nil, err
	}
	switch value := raw.(type) {
	case []uint16:
		return value, nil
	case []uint32:
		parts := make([]uint16, len(value))
		for i, v := range value {
			if v > 0xffff {
				return nil, fmt.Errorf("parts list entry out of range: %d", v)
			}
			parts[i] = uint16(v)
		}
		return parts, nil
	case []int:
		parts := make([]uint16, len(value))
		for i, v := range value {
			if v < 0 || v > 0xffff {
				return nil, fmt.Errorf("parts list entry out of range: %d", v)
			}
			parts[i] = uint16(v)
		}
		return parts, nil
	case []any:
		parts := make([]uint16, 0, len(value))
		for _, entry := range value {
			switch v := entry.(type) {
			case uint16:
				parts = append(parts, v)
			case uint32:
				if v > 0xffff {
					return nil, fmt.Errorf("parts list entry out of range: %d", v)
				}
				parts = append(parts, uint16(v))
			case int:
				if v < 0 || v > 0xffff {
					return nil, fmt.Errorf("parts list entry out of range: %d", v)
				}
				parts = append(parts, uint16(v))
			default:
				return nil, fmt.Errorf("unsupported parts list entry type %T", entry)
			}
		}
		return parts, nil
	default:
		return nil, fmt.Errorf("unsupported parts list type %T", raw)
	}
}

func readBridgedDevice(ctx context.Context, ctrl Controller, nodeID uint64, endpoint uint16) (BridgedDevice, bool, error) {
	uniqueID, err := readStringAttribute(ctx, ctrl, nodeID, endpoint, bridgedDeviceBasicInfoClusterID, bridgedUniqueIDAttrID)
	if err != nil {
		if errors.Is(err, errAttributeUnavailable) {
			return BridgedDevice{}, false, nil
		}
		return BridgedDevice{}, false, err
	}
	nodeLabel, err := readStringAttribute(ctx, ctrl, nodeID, endpoint, bridgedDeviceBasicInfoClusterID, bridgedNodeLabelAttrID)
	if err != nil && !errors.Is(err, errAttributeUnavailable) {
		return BridgedDevice{}, false, err
	}
	reachable, err := readBoolAttribute(ctx, ctrl, nodeID, endpoint, bridgedDeviceBasicInfoClusterID, bridgedReachableAttrID)
	if err != nil && !errors.Is(err, errAttributeUnavailable) {
		return BridgedDevice{}, false, err
	}

	device := BridgedDevice{
		NodeID:    nodeID,
		Endpoint:  endpoint,
		UniqueID:  uniqueID,
		NodeLabel: nodeLabel,
		Reachable: reachable,
	}
	return device, true, nil
}

var errAttributeUnavailable = errors.New("attribute unavailable")

func readStringAttribute(ctx context.Context, ctrl Controller, nodeID uint64, endpoint uint16, clusterID uint32, attrID uint32) (string, error) {
	raw, err := ctrl.ReadAttribute(ctx, nodeID, endpoint, clusterID, attrID)
	if err != nil {
		return "", err
	}
	if raw == nil {
		return "", errAttributeUnavailable
	}
	switch value := raw.(type) {
	case string:
		return value, nil
	case []byte:
		return string(value), nil
	default:
		return "", fmt.Errorf("unsupported string attribute type %T", raw)
	}
}

func readBoolAttribute(ctx context.Context, ctrl Controller, nodeID uint64, endpoint uint16, clusterID uint32, attrID uint32) (*bool, error) {
	raw, err := ctrl.ReadAttribute(ctx, nodeID, endpoint, clusterID, attrID)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, errAttributeUnavailable
	}
	switch value := raw.(type) {
	case bool:
		return &value, nil
	case *bool:
		return value, nil
	default:
		return nil, fmt.Errorf("unsupported bool attribute type %T", raw)
	}
}
