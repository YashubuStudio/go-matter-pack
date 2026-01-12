package store

import (
	"time"

	"github.com/cybergarage/go-matter-pack/internal/mattermodel"
)

// Registry holds stable device identifiers keyed by UniqueID.
type Registry struct {
	HubNodeID uint64                  `json:"hub_node_id"`
	UpdatedAt time.Time               `json:"updated_at"`
	Devices   map[string]DeviceRecord `json:"devices"`
}

// DeviceRecord stores the latest known state for a bridged device.
type DeviceRecord struct {
	UniqueID         string    `json:"unique_id"`
	NodeID           uint64    `json:"node_id"`
	Endpoint         uint16    `json:"endpoint"`
	LastSeenAt       time.Time `json:"last_seen_at"`
	LastSeenEndpoint uint16    `json:"last_seen_endpoint"`
	Label            string    `json:"label,omitempty"`
	NodeLabel        string    `json:"node_label,omitempty"`
	Reachable        *bool     `json:"reachable,omitempty"`
	Missing          bool      `json:"missing,omitempty"`
}

// NewRegistry initializes an empty registry.
func NewRegistry() *Registry {
	return &Registry{Devices: make(map[string]DeviceRecord)}
}

// Find returns a device record by UniqueID.
func (r *Registry) Find(uniqueID string) (DeviceRecord, bool) {
	if r == nil || r.Devices == nil {
		return DeviceRecord{}, false
	}
	record, ok := r.Devices[uniqueID]
	return record, ok
}

// ApplyScan updates the registry using the latest bridged device scan.
func (r *Registry) ApplyScan(now time.Time, nodeID uint64, devices []mattermodel.BridgedDevice) {
	if r.Devices == nil {
		r.Devices = make(map[string]DeviceRecord)
	}
	seen := make(map[string]struct{}, len(devices))
	for _, dev := range devices {
		if dev.UniqueID == "" {
			continue
		}
		seen[dev.UniqueID] = struct{}{}
		record := r.Devices[dev.UniqueID]
		record.UniqueID = dev.UniqueID
		record.NodeID = nodeID
		record.Endpoint = dev.Endpoint
		record.LastSeenAt = now
		record.LastSeenEndpoint = dev.Endpoint
		record.NodeLabel = dev.NodeLabel
		record.Reachable = dev.Reachable
		record.Missing = false
		if record.Label == "" && dev.NodeLabel != "" {
			record.Label = dev.NodeLabel
		}
		r.Devices[dev.UniqueID] = record
	}

	for uniqueID, record := range r.Devices {
		if _, ok := seen[uniqueID]; !ok {
			record.Missing = true
			r.Devices[uniqueID] = record
		}
	}

	r.HubNodeID = nodeID
	r.UpdatedAt = now
}

// SetLabel fixes the user label for a device.
func (r *Registry) SetLabel(uniqueID, label string) bool {
	record, ok := r.Devices[uniqueID]
	if !ok {
		return false
	}
	record.Label = label
	r.Devices[uniqueID] = record
	return true
}

// Remove deletes a device record from the registry.
func (r *Registry) Remove(uniqueID string) {
	delete(r.Devices, uniqueID)
}
